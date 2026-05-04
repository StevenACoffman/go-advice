Validating data in Go [#](#validating-data-in-go)
=================================================

Mat Ryer’s approach to [validating data in Go](https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years/#validating-data) is simple and elegant. I have been using it extensively while building [WhyNotLog](https://whynotlog.com). The idea is to define a `Validator` interface and any type that implements this interface is validation-ready:

    type Validator interface {
        Validate() map[string]string
    }
    

The `Validate` method returns a map of field names to problem descriptions. The map is a bit unwieldy, so I like returning a `Problems` type instead:

    type Problems map[string]string
    

We can add syntactic sugar to make it more convenient to use:

    func (p Problems) Add(field, msg string) {
        p[field] = msg
    }
    

We also want a `String` method to turn problems into a human-readable string representation:

    func (p Problems) String() string {
        if len(p) == 0 {
            return ""
        }
    
        msgs := make([]string, 0, len(p))
        for field, msg := range p {
            msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
        }
        return strings.Join(msgs, "; ")
    }
    

Implementing the interface [#](#implementing-the-interface)
-----------------------------------------------------------

With all that tooling in place, consider this `TaskCreateRequest` type. Callers use this type to register a task that is meant to be run periodically.

    type TaskCreateRequest struct {
        Name     string
        Run      func(ctx context.Context) error
        Interval time.Duration
        Deadline time.Duration
    }
    

A valid value of this type must satisfy a number of constraints:

*   The name must not be empty.
*   The run function must not be `nil`.
*   Both the interval and deadline must be greater than 0.
*   The deadline must be less than the interval, otherwise two instances of the task can run simultaneously.

We implement these constraints by having the `TaskCreateRequest` type implement the `Validator` interface:

    func (r *TaskCreateRequest) Validate() validation.Problems {
        problems := make(validation.Problems)
    
        if r.Name == "" {
            problems.Add("Name", "is required")
        }
        if r.Run == nil {
            problems.Add("Run", "is required")
        }
        if r.Interval <= 0 {
            problems.Add("Interval", "must be greater than 0")
        }
        if r.Deadline <= 0 {
            problems.Add("Deadline", "must be greater than 0")
        }
        if r.Deadline >= r.Interval {
            problems.Add("Deadline", "must be less than Interval")
        }
    
        return problems
    }
    

The `Problems` type is defined in the `validation` package, which gives us the nicely descriptive qualified name `validation.Problems`.

Validating types [#](#validating-types)
---------------------------------------

Calling `Validate` directly is a bit cumbersome because `validation.Problems` does not implement the `error` interface and callers shouldn’t have to bother with the custom type. Instead, we define a `Check` function – also in the `validation` package – that calls `Validate` and returns an error if the given type has one or more problems. We use some reflection magic to get the type’s name.

    func Check(v Validator) error {
        if v == nil {
            return errors.New("value is nil, cannot validate")
        }
    
        problems := v.Validate()
        if len(problems) == 0 {
            return nil
        }
    
        return fmt.Errorf("validation failed for %s: %s",
            typeName(v),
            problems.String(),
        )
    }
    
    func typeName(v any) string {
        t := reflect.TypeOf(v)
        if t == nil {
            return "value"
        }
        // Keep on dereferencing the pointer until we get to the concrete type.
        for t.Kind() == reflect.Pointer {
            t = t.Elem()
        }
        if t.Name() != "" {
            return t.Name()
        }
        return t.String()
    }
    

The task registration code now only has to call `Check` and handle the error in a Go-idiomatic way:

    func RegisterTask(req *TaskCreateRequest) error {
        if err := validation.Check(req); err != nil {
            return err
        }
    
        ...
    }
    

If we provide an invalid task creation request, like this one:

    if err := RegisterTask(&TaskCreateRequest{
        // Name and Deadline are missing!
        Run: func(ctx context.Context) error {
            return nil
        },
        Interval: time.Second,
    }); err != nil {
        log.Fatal(err)
    }
    

…we end up with the following error:

    2026/04/16 16:56:31 validation failed for TaskCreateRequest: Deadline: must be greater than 0; Name: is required
    

Succinct and actionable.

Where should we validate? [#](#where-should-we-validate)
--------------------------------------------------------

Go programs that provide a Web API often follow the controller/service/repository pattern. Controllers implement an access layer that allows clients to interact with the system, typically by exposing an HTTP API, or gRPC, or even just using `stdin` and `stdout`. Services implement business logic. Repositories implement persistence, typically by interacting with a database, or by implementing an in-memory data store, which can be useful for testing.

In a controller/service/repository pattern, the service is responsible for validating whatever enters its boundaries. Every service method that allows for the creation of a new resource (or the modification of an existing one) validates incoming data.

I find it convenient to have validation code right next to the corresponding type definition. Whenever I find that validation logic is incomplete (bad data makes it all the way to the database) or overzealous (good data is rejected prematurely), we instantly know where to look and adjust the logic.