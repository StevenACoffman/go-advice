Dependency Composition
======================

_Based on frustrations with conventional framework-based dependency injection, I have adopted a composition strategy that uses closures over dependency structs to inject context into modules. When combined with test-driven development as the design process and a focus on functions over types, modules can be kept clear, clean, and mostly free of unintended coupling. This is an adaptation of Daniel Somerfield's original article for Go._

23 May 2023 · Go adaptation

* * *

[![Photo of Daniel Somerfield](dependency-composition/daniel.jpg)](https://github.com/danielsomerfield)

[Daniel Somerfield](https://github.com/danielsomerfield)

As a software developer and sometimes manager with over two decades of software industry experience, I am committed to designing systems that help people live fulfilled, meaningful lives with access to the basic human needs of physical health and psychological safety. At Measurabl, I focus on developing products that delight our customers, build our business, and make significant progress toward reducing the negative climate impact of building operations.

[object collaboration design](/tags/object%20collaboration%20design.html)

[application architecture](/tags/application%20architecture.html)

Contents
--------

*   [Origin Story](#OriginStory)
*   [Dependency injection is a means, not an end](#DependencyInjectionIsAMeansNotAnEnd)
*   [The problem description](#TheProblemDescription)
*   [Building a solution](#BuildingASolution)
*   [On to the controller](#OnToTheController)
*   [Digging into the domain](#DiggingIntoTheDomain)
*   [Beginning to wire it up](#BeginningToWireItUp)
*   [Finding and fixing a problem](#FindingAndFixingAProblem)
*   [Reaching out to the repository layer](#ReachingOutToTheRepositoryLayer)
*   [The last mile: implementing domain layer dependencies](#TheLastMileImplementingDomainLayerDependencies)
*   [If it looks like a duck](#IfItLooksLikeADuck)
*   [In summary](#InSummary)

### Sidebars

*   [What do I mean by "incidental coupling"?](#WhatDoIMeanByincidentalCoupling)
*   [On mocks, stubs, and frameworks](#OnMocksStubsAndFrameworks)
*   [Rating constants in Go](#RatingConstantsInGo)
*   [Method values in Go](#MethodValuesInGo)

* * *

Origin Story
------------

It all started a few years ago when members of one of my teams asked, "what pattern should we adopt for [dependency injection](https://www.martinfowler.com/articles/injection.html#FormsOfDependencyInjection) (DI)"? The team's stack was Go, not one I was terribly familiar with, so I encouraged them to work it out for themselves. I was disappointed to learn some time later that the team had decided, in effect, not to decide, leaving behind a plethora of patterns for wiring modules together. Some developers used constructor injection, others manual dependency injection in root modules, and some global variables in `init()` functions.

The results were less than ideal: a hodgepodge of patterns assembled in different ways, each requiring a very different approach to testing. Some modules were unit testable, others lacked entry points for testing, so simple logic required complex HTTP-aware scaffolding to exercise basic functionality. Most critically, changes in one part of the codebase sometimes caused broken contracts in unrelated areas. Some modules were interdependent across packages; others had completely flat collections of files with no distinction between subdomains.

With the benefit of hindsight, I continued to think about that original decision: what DI pattern should we have picked. Ultimately I came to a conclusion: that was the wrong question.

Dependency injection is a means, not an end
-------------------------------------------

In retrospect, I should have guided the team towards asking a different question: what are the desired qualities of our codebase, and what approaches should we use to achieve them? I wish I had advocated for the following:

*   discrete modules with minimal incidental coupling, even at the cost of some duplicate types
*   business logic that is kept from intermingling with code that manages the transport, like HTTP handlers or gRPC implementations
*   business logic tests that are not transport-aware or have complex scaffolding
*   tests that do not break when new fields are added to types
*   very few types exposed outside of their packages, and even fewer types exposed outside of the directories they inhabit.

What do I mean by "incidental coupling"?
----------------------------------------

Often when I hear developers talk about coupling, it's either to extol the virtues of "loose" coupling or to warn against the evils of it when it's "tight." Often absent from the conversation are nuances of context. The context tells us whether the relationship between these parts is appropriate and whether it reflects a reality of the domain that is valuable to model. In this article, I use "incidentally coupling" to describe relationships between modules that increase code fragility without significant benefits, like enforcing a constraint necessary to the domain.

Over the last few years, I've settled on an approach that leads a developer who adopts it toward these qualities. Having come from a [Test-Driven Development](https://martinfowler.com/bliki/TestDrivenDevelopment.html) (TDD) background, I naturally start there. TDD encourages incrementalism but I wanted to go even further, so I have taken a minimalist "function-first" approach to module composition. Rather than continuing to describe the process, I will demonstrate it. What follows is an example web service built on a relatively simple architecture wherein a handler package calls domain logic which in turn calls repository functions in the persistence layer.

The problem description
-----------------------

Imagine a user story that looks something like this:

As a registered user of RateMyMeal and a would-be restaurant patron who doesn't know what's available, I would like to be provided with a ranked set of recommended restaurants in my region based on other patron ratings.

**Acceptance Criteria**

*   The restaurant list is ranked from the most to the least recommended.
*   The rating process includes the following potential rating levels:

*   excellent (2)
*   above average (1)
*   average (0)
*   below average (-1)
*   terrible (-2).

*   The overall rating is the sum of all individual ratings.
*   Users considered "trusted" get a 4X multiplier on their rating.
*   The user must specify a city to limit the scope of the returned restaurant.

Building a solution
-------------------

I have been tasked with building a REST service using Go and PostgreSQL. I start by building a very coarse integration as a [walking skeleton](https://wiki.c2.com/?WalkingSkeleton) that defines the boundaries of the problem I wish to solve. This test uses as much of the underlying infrastructure as possible. If I use any stubs, it's for third-party cloud providers or other services that can't be run locally. Even then, I use server stubs, so I can use real SDKs or network clients. This becomes my acceptance test for the task at hand, keeping me focused. I will only cover one "happy path" that exercises the basic functionality since the test will be time-consuming to build robustly. I'll find less costly ways to test edge cases. For the sake of the article, I assume that I have a skeletal database structure that I can modify if required.

![](dependency-composition/db.png)

Tests generally have a `given/when/then` structure: a set of given conditions, a participating action, and a verified result. I prefer to start at `when/then` and back into the `given` to help me focus the problem I'm trying to solve.

"**When** I call my recommendation endpoint, **then** I expect to get an OK response and a payload with the top-rated restaurants based on our ratings algorithm". In code that could be:

test/e2e\_integration\_test.go…

    func TestRestaurantsEndpoint(t *testing.T) {
        t.Run("ranks by the recommendation heuristic", func(t *testing.T) {
            resp, err := http.Get(    ➀
                "http://localhost:3000/vancouverbc/restaurants/recommended",
            )
            if err != nil {
                t.Fatalf("request failed: %v", err)
            }
            defer resp.Body.Close()

            if resp.StatusCode != http.StatusOK {
                t.Fatalf("expected 200, got %d", resp.StatusCode)
            }

            var data ResponsePayload
            if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
                t.Fatalf("decode failed: %v", err)
            }

            ids := make([]string, len(data.Restaurants))
            for i, r := range data.Restaurants {
                ids[i] = r.ID
            }
            if !slices.Equal(ids, []string{"cafegloucesterid", "burgerkingid"}) {    ➁
                t.Errorf("unexpected order: %v", ids)
            }
        })
    }

    type ResponsePayload struct {
        Restaurants []struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        } `json:"restaurants"`
    }

There are a couple of details worth calling out:

1.  `net/http` is Go's standard library HTTP client, no third-party dependency required. Go's type system ensures fields accessed on `ResponsePayload` are checked at compile time; my assertions verify the runtime response actually contains those fields.
2.  Rather than checking the entire contents of the returned restaurants, I only check their ids. This small detail is deliberate. If I check the contents of the entire object, my test becomes fragile, breaking if I add a new field. I want to write a test that will accommodate the natural evolution of my code while at the same time verifying the specific condition I'm interested in: the order of the restaurant listing.

Without my `given` conditions, this test isn't very valuable, so I add them next.

test/e2e\_integration\_test.go…

    func TestRestaurantsEndpoint(t *testing.T) {
        var srv *http.Server
        var db *sql.DB

        users := []struct {
            ID      string
            Name    string
            Trusted bool
        }{
            {"u1", "User1", true},
            {"u2", "User2", false},
            {"u3", "User3", false},
        }

        restaurants := []struct {
            ID   string
            Name string
        }{
            {"cafegloucesterid", "Cafe Gloucester"},
            {"burgerkingid", "Burger King"},
        }

        ratingsByUser := [][]any{
            {"rating1", users[0], restaurants[0], "EXCELLENT"},
            {"rating2", users[1], restaurants[0], "TERRIBLE"},
            {"rating3", users[2], restaurants[0], "AVERAGE"},
            {"rating4", users[2], restaurants[1], "ABOVE_AVERAGE"},
        }

        t.Cleanup(func() {    ➀
            if srv != nil {
                _ = srv.Shutdown(context.Background())
            }
            if db != nil {
                _ = db.Close()
            }
        })

        // GIVEN
        var err error
        db, err = startTestDatabase(t)    ➁
        if err != nil {
            t.Fatalf("database setup: %v", err)
        }

        for _, u := range users {
            if err := createUser(db, u); err != nil {
                t.Fatalf("createUser: %v", err)
            }
        }
        for _, r := range restaurants {
            if err := createRestaurant(db, r); err != nil {
                t.Fatalf("createRestaurant: %v", err)
            }
        }
        for _, rating := range ratingsByUser {
            if err := createRatingByUserForRestaurant(db, rating); err != nil {
                t.Fatalf("createRating: %v", err)
            }
        }

        srv, err = startServer(db)
        if err != nil {
            t.Fatalf("server start: %v", err)
        }

        t.Run("ranks by the recommendation heuristic", func(t *testing.T) {
            // .. snip
        })
    }

My `given` conditions are set up before the subtest runs, and `t.Cleanup` handles teardown regardless of test outcome. Go 1.21+ provides `t.Cleanup` as the idiomatic replacement for setup/teardown hooks; it is scoped to the test, runs even on failure, and composes cleanly when subtests register their own cleanup. You'll notice explicit `error` returns throughout. Go's convention is that anything IO-bound — database calls, file reads, network requests — returns an `error` as its last return value. Unlike the Node.js pattern of defaulting every contract to `Promise<T>` because making a sync function async later is painful, Go's convention is simpler: just return `(T, error)`. Concurrency is layered in explicitly (via goroutines and `context.Context`) only where it is actually needed.

I've intentionally deferred creating explicit types for the users and restaurants, acknowledging I don't know what they look like yet. Unlike TypeScript, Go does not have structural typing for struct literals — but Go _interfaces_ are structural, satisfied implicitly by any type that has the right methods. This means I can continue to defer exposing shared types while still getting compile-time contract checking as my module APIs solidify. We'll come back to this distinction in detail later.

At this point, I have a shell of a test with test dependencies missing. The next stage is to flesh out those dependencies by first building stub functions to get the test to compile and then implementing these helper functions. That is a non-trivial amount of work, but it's also highly contextual and out of the scope of this article. Suffice it to say that it will generally consist of:

*   starting up dependent services, such as databases. I generally use [testcontainers-go](https://golang.testcontainers.org/) to run dockerized services, but these could also be network fakes or in-memory components, whatever you prefer.
*   fill in the `create...` functions to pre-construct the entities required for the test. In the case of this example, these are SQL `INSERT`s.
*   start up the service itself, at this point a simple stub. We'll dig a little more into the service initialization since it's germane to the discussion of composition.

If you are interested in how the test dependencies are initialized, you can see [the results](https://github.com/danielsomerfield/function-first-composition-example-go) in the GitHub repo.

Before moving on, I run the test to make sure it fails as I would expect. Because I have not yet implemented my server `start`, I expect to receive a connection refused error when making my HTTP request. With that confirmed, I disable my big integration test, since it's not going to pass for a while, and commit.

On to the controller
--------------------

I generally build from the outside in, so my next step is to address the main HTTP handling function. First, I'll build a handler unit test. I start with something that ensures an empty 200 response with expected headers:

On mocks, stubs, and frameworks
-------------------------------

Go's standard library includes `net/http/httptest`, which provides `httptest.NewRequest` and `httptest.NewRecorder` for testing HTTP handlers without spinning up a real server. These are real request and response objects, not custom stubs. Unlike the TypeScript version of this article — which required hand-rolling `stubRequest()` and `stubResponse()` helpers because Express request/response types are not easily constructable in tests — Go's stdlib makes handler testing first-class. No framework needed, no bespoke test infrastructure required.

test/restaurantratings/controller\_test.go…

    func TestRatingsController(t *testing.T) {
        t.Run("provides a JSON response with ratings", func(t *testing.T) {
            handler := controller.NewTopRatedHandler(controller.Dependencies{})    ➀
            req := httptest.NewRequest(http.MethodGet, "/vancouverbc/restaurants/recommended", nil)
            w := httptest.NewRecorder()

            handler(w, req)

            res := w.Result()
            if res.StatusCode != http.StatusOK {
                t.Errorf("expected 200, got %d", res.StatusCode)
            }
            if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
                t.Errorf("expected json content-type, got %q", ct)
            }
        })
    }

I've already started to do a little design work that will result in the highly decoupled modules I promised. Most of the code is fairly typical test scaffolding, but if you look closely at the highlighted function call it might strike you as unusual.

This small detail is the first step toward the closure-over-dependencies pattern, where a factory function (`NewTopRatedHandler`) captures its dependencies and returns a function that uses them. In the coming paragraphs, I'll demonstrate how it becomes the foundation upon which the compositional approach is built.

Next, I build out the stub of the unit under test, this time the handler, and run it to ensure my test is operating as expected:

src/restaurantratings/controller.go…

    func NewTopRatedHandler(deps Dependencies) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {}
    }

My test expects a 200, but I get no calls to write a status code, so the test fails. A minor tweak and it's passing:

src/restaurantratings/controller.go…

    func NewTopRatedHandler(deps Dependencies) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            _ = json.NewEncoder(w).Encode(struct{}{})
        }
    }

I commit and move on to fleshing out the test for the expected payload. I don't yet know exactly how I will handle the data access or algorithmic part of this application, but I know that I would like to delegate, leaving this module to nothing but translate between the HTTP protocol and the domain. I also know what I want from the delegate. Specifically, I want it to load the top-rated restaurants, whatever they are and wherever they come from, so I create a `Dependencies` stub that has a function field to return the top restaurants. This becomes a field in my `Dependencies` struct.

test/restaurantratings/controller\_test.go…

    type Restaurant struct{ ID, Name string }

    type RestaurantResponseBody struct {
        Restaurants []Restaurant `json:"restaurants"`
    }

    var vancouverRestaurants = []Restaurant{
        {ID: "cafegloucesterid", Name: "Cafe Gloucester"},
        {ID: "baravignonid", Name: "Bar Avignon"},
    }

    var topRestaurantsByCity = map[string][]Restaurant{
        "vancouverbc": vancouverRestaurants,
    }

    depsStub := controller.Dependencies{
        GetTopRestaurants: func(ctx context.Context, city string) ([]Restaurant, error) {    ➀
            return topRestaurantsByCity[city], nil
        },
    }

    req := httptest.NewRequest(
        http.MethodGet,
        "/vancouverbc/restaurants/recommended",
        nil,
    )
    req = req.WithContext(r.Context())    // carry real context through
    req.SetPathValue("city", "vancouverbc")    ➁
    w := httptest.NewRecorder()

    controller.NewTopRatedHandler(depsStub)(w, req)

    res := w.Result()
    // ... status/content-type checks ...
    var body RestaurantResponseBody
    _ = json.NewDecoder(res.Body).Decode(&body)
    if len(body.Restaurants) != 2 {
        t.Fatalf("expected 2 restaurants, got %d", len(body.Restaurants))
    }
    // check IDs only — adding fields later won't break this test
    if body.Restaurants[0].ID != "cafegloucesterid" {
        t.Errorf("wrong first restaurant: %v", body.Restaurants[0].ID)
    }

1.  `GetTopRestaurants` is a function field in the `Dependencies` struct. Its signature takes `context.Context` as its first argument. `context.Context` is how Go propagates deadlines, cancellation signals, and request-scoped values. Passing `ctx` first is a universal Go convention; it plays the role that async-by-default plays in Node.js, but opt-in and explicit.
2.  `r.SetPathValue` sets named path parameters; available from Go 1.22's enhanced `net/http` router.

With so little information on how `GetTopRestaurants` is implemented, how do I stub it? I know enough to design a basic client view of the contract I've created implicitly in my dependencies stub: a simple function that takes a context and a city, and returns a slice of restaurants or an error. This contract might be fulfilled by a simple static function, a method on an object instance, or a stub, as in the test above. This module doesn't know, doesn't care, and doesn't have to. It is exposed to the minimum it needs to do its job, nothing more.

src/restaurantratings/controller.go…

    type Restaurant struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }

    type Dependencies struct {
        GetTopRestaurants func(ctx context.Context, city string) ([]Restaurant, error)
    }

    func NewTopRatedHandler(deps Dependencies) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            city := r.PathValue("city")
            restaurants, err := deps.GetTopRestaurants(r.Context(), city)
            if err != nil {
                http.Error(w, "internal server error", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            _ = json.NewEncoder(w).Encode(struct {
                Restaurants []Restaurant `json:"restaurants"`
            }{restaurants})
        }
    }

For those who like to visualize these things, we can visualize the production code so far as the handler function that requires something that implements the `GetTopRestaurants` contract using a [ball and socket](https://martinfowler.com/bliki/BallAndSocket.html) notation.

controller.go

handler()

GetTopRestaurants()

The tests create this function and a stub for the required function. I can show this by using a different colour for the tests, and the socket notation to show implementation of an interface.

controller.go controller\_test.go

handler()

GetTop
Restaurants()

test

GetTopRestaurants()

This `controller` package is brittle at this point, so I'll need to flesh out my tests to cover alternative code paths and edge cases, but that's a bit beyond the scope of the article. If you're interested in seeing a more thorough test and the resulting controller module, both are available in the [GitHub repo](https://github.com/danielsomerfield/function-first-composition-example-go).

Digging into the domain
-----------------------

At this stage, I have a handler that requires a function that doesn't exist. My next step is to provide a module that can fulfill the `GetTopRestaurants` contract. I'll start that process by writing a big clumsy unit test and refactor it for clarity later. It is only at this point I start thinking about how to implement the contract I have previously established. I go back to my original acceptance criteria and try to minimally design my module.

test/restaurantratings/toprated\_test.go…

    func TestTopRatedRestaurantList(t *testing.T) {
        t.Run("is calculated from our proprietary ratings algorithm", func(t *testing.T) {
            ratings := []toprated.RatingsByRestaurant{    ➃
                {
                    RestaurantID: "restaurant1",
                    Ratings: []toprated.RestaurantRating{
                        {Rating: toprated.Excellent},
                    },
                },
                {
                    RestaurantID: "restaurant2",
                    Ratings: []toprated.RestaurantRating{
                        {Rating: toprated.Average},
                    },
                },
            }

            ratingsByCity := map[string][]toprated.RatingsByRestaurant{
                "vancouverbc": ratings,
            }

            findRatingsByRestaurantStub := func(ctx context.Context, city string) ([]toprated.RatingsByRestaurant, error) {    ➀
                return ratingsByCity[city], nil
            }

            calculateRatingForRestaurantStub := func(r toprated.RatingsByRestaurant) int {    ➁
                // predictable stub: don't know the algorithm yet
                switch r.RestaurantID {
                case "restaurant1":
                    return 10
                case "restaurant2":
                    return 5
                default:
                    panic("unknown restaurant: " + r.RestaurantID)
                }
            }

            deps := toprated.Dependencies{    ➂
                FindRatingsByRestaurant:      findRatingsByRestaurantStub,
                CalculateRatingForRestaurant: calculateRatingForRestaurantStub,
            }

            getTopRated := toprated.New(deps)
            topRestaurants, err := getTopRated(context.Background(), "vancouverbc")
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if len(topRestaurants) != 2 {
                t.Fatalf("expected 2 restaurants, got %d", len(topRestaurants))
            }
            if topRestaurants[0].ID != "restaurant1" {
                t.Errorf("expected restaurant1 first, got %v", topRestaurants[0].ID)
            }
            if topRestaurants[1].ID != "restaurant2" {
                t.Errorf("expected restaurant2 second, got %v", topRestaurants[1].ID)
            }
        })
    }

    type Restaurant struct {    // local to test file, not yet promoted
        ID string
    }

I have introduced a lot of new concepts into the domain at this point, so I'll take them one at a time:

1.  I need a "finder" that returns a set of ratings for each restaurant. I'll start by stubbing that out. Its signature is `func(ctx context.Context, city string) ([]RatingsByRestaurant, error)` — the canonical Go async contract.
2.  The acceptance criteria provide the algorithm that will drive the overall rating, but I choose to ignore that for now and say that, _somehow_, this group of ratings will provide the overall restaurant rating as a numeric value.
3.  For this module to function it will rely on two new concepts: finding the ratings of a restaurant, and given that set of ratings, producing an overall rating. I create another `Dependencies` struct that includes the two stubbed functions with naive, predictable stub implementations to keep me moving forward.
4.  The `RatingsByRestaurant` and `RestaurantRating` types I define in the test communicate intent. In Go, types defined in a `_test.go` file are only visible to test code in the same package; this enforces the "don't export until you're certain" discipline mechanically. Unlike TypeScript, where the test compiler doesn't prevent production code from importing test types, Go's build system simply won't compile it.

Rating constants in Go
-----------------------

Rather than a TypeScript `const` object trick, I define a named integer type and use a typed `const` block. This gives each constant a distinct type, preventing accidental mixing with plain `int`. For parsing rating strings from external data (the database or JSON), I maintain an explicit map:

    type Rating int

    const (
        Excellent    Rating = 2
        AboveAverage Rating = 1
        Average      Rating = 0
        BelowAverage Rating = -1
        Terrible     Rating = -2
    )

    var RatingByName = map[string]Rating{
        "EXCELLENT":     Excellent,
        "ABOVE_AVERAGE": AboveAverage,
        "AVERAGE":       Average,
        "BELOW_AVERAGE": BelowAverage,
        "TERRIBLE":      Terrible,
    }

The typed constant approach gives you behaviour via methods on the `Rating` type (e.g., `func (r Rating) String() string`) without the indirection of a lookup table for the common case. The `map` is reserved for the boundary where string values enter from the outside world.

Once the basic structure of my test is in place, I try to make it compile with a minimalist implementation.

src/restaurantratings/toprated.go…

    type Dependencies struct{}    // empty for now

    func New(deps Dependencies) func(ctx context.Context, city string) ([]Restaurant, error) {    ➀
        return func(ctx context.Context, city string) ([]Restaurant, error) {
            return nil, nil
        }
    }

    type Restaurant struct {    ➁
        ID string
    }

    type Rating int    ➂

    const (
        Excellent    Rating = 2
        AboveAverage Rating = 1
        Average      Rating = 0
        BelowAverage Rating = -1
        Terrible     Rating = -2
    )

1.  I use my closure-over-dependencies factory pattern, capturing the `Dependencies` struct and returning a function. The test will fail, of course, but seeing it fail in the way I expect builds my confidence that it is sound.
2.  As I begin implementing the module under test, I identify some domain objects that should be promoted to production code. In particular, I move the direct dependencies into the module under test. Anything that isn't a direct dependency, I leave where it is in test code.
3.  I also make one anticipatory move: I extract the `Rating` type into production code. I feel comfortable doing so because it is a universal and explicit domain concept. The values were specifically called out in the acceptance criteria, which says to me that couplings are less likely to be incidental.

Notice that the types I define or move into the production code are _not_ exported from their packages. That is a deliberate choice, one I'll discuss in more depth later. Suffice it to say, I have yet to decide whether I want other packages binding to these types, creating more couplings that might prove to be undesirable. In Go, unexported (lowercase) identifiers are enforced by the compiler — no convention or linter needed.

Now, I finish the implementation of the `toprated.go` module.

src/restaurantratings/toprated.go…

    type Dependencies struct {    ➀
        FindRatingsByRestaurant      func(ctx context.Context, city string) ([]RatingsByRestaurant, error)
        CalculateRatingForRestaurant func(ratings RatingsByRestaurant) int
    }

    type overallRating struct {    ➁
        restaurantID string
        rating       int
    }

    type RestaurantRating struct {    ➂
        Rating Rating
    }

    type RatingsByRestaurant struct {
        RestaurantID string
        Ratings      []RestaurantRating
    }

    func New(deps Dependencies) func(ctx context.Context, city string) ([]Restaurant, error) {    ➃
        calculateRatings := func(
            ratingsByRestaurant []RatingsByRestaurant,
            calc func(RatingsByRestaurant) int,
        ) []overallRating {
            result := make([]overallRating, len(ratingsByRestaurant))
            for i, r := range ratingsByRestaurant {
                result[i] = overallRating{
                    restaurantID: r.RestaurantID,
                    rating:       calc(r),
                }
            }
            return result
        }

        return func(ctx context.Context, city string) ([]Restaurant, error) {
            ratingsByRestaurant, err := deps.FindRatingsByRestaurant(ctx, city)
            if err != nil {
                return nil, fmt.Errorf("find ratings: %w", err)
            }

            overallRatings := calculateRatings(ratingsByRestaurant, deps.CalculateRatingForRestaurant)
            sortByOverallRating(overallRatings)

            restaurants := make([]Restaurant, len(overallRatings))
            for i, r := range overallRatings {
                restaurants[i] = Restaurant{ID: r.restaurantID}
            }
            return restaurants, nil
        }
    }

    func sortByOverallRating(ratings []overallRating) {
        sort.Slice(ratings, func(i, j int) bool {
            return ratings[i].rating > ratings[j].rating
        })
    }

    // SNIP ...

Having done so, I have

1.  filled out the `Dependencies` struct I modeled in my unit test
2.  introduced the `overallRating` type to capture the domain concept. Lowercase because it's internal to this package; there's no reason yet to share it. Go keeps unexported types entirely invisible to callers.
3.  extracted a couple of types from the test that are now direct dependencies of my `toprated` module
4.  completed the simple logic of the primary function returned by the factory.

The dependencies between the main production code functions look like this

controller.go toprated.go

handler()

topRated()

GetTopRestaurants() FindRatingsByRestaurant() CalculateRatings
ForRestaurants()

When including the stubs provided by the test, it looks like this

controller.go toprated.go controller\_test.go

handler()

topRated()

calculateRatingFor
RestaurantStub()

findRatingsBy
RestaurantStub

test

GetTopRestaurants() FindRatingsByRestaurant() CalculateRatings
ForRestaurants()

With this implementation complete (for now), I have a passing test for my main domain function and one for my handler. They are entirely decoupled. So much so, in fact, that I feel the need to prove to myself that they will work together. It's time to start composing the units and building toward a larger whole.

Beginning to wire it up
-----------------------

At this point, I have a decision to make. If I'm building something relatively straightforward, I might choose to dispense with a test-driven approach when integrating the modules, but in this case, I'm going to continue down the TDD path for two reasons:

*   I want to focus on the design of the integrations between modules, and writing a test is a good tool for doing so.
*   There are still several modules to be implemented before I can use my original acceptance test as validation. If I wait to integrate them until then, I might have a lot to untangle if some of my underlying assumptions are flawed.

If my first acceptance test is a boulder and my unit tests are pebbles, then this first integration test would be a fist-sized rock: a chunky test exercising the call path from the handler into the first layer of domain functions, providing test doubles for anything beyond that layer. At least that is how it will start. I might continue integrating subsequent layers of the architecture as I go. I also might decide to throw the test away if it loses its utility or is getting in my way.

After initial implementation, the test will validate little more than that I've wired the routes correctly, but will soon cover calls into the domain layer and validate that the responses are encoded as expected.

test/restaurantratings/controller\_integration\_test.go…

    func TestControllerTopRatedHandler(t *testing.T) {
        t.Run("delegates to the domain top rated logic", func(t *testing.T) {
            returnedRestaurants := []controller.Restaurant{
                {ID: "r1", Name: "restaurant1"},
                {ID: "r2", Name: "restaurant2"},
            }

            topRatedStub := func(ctx context.Context, city string) ([]controller.Restaurant, error) {
                return returnedRestaurants, nil
            }

            mux := http.NewServeMux()
            mux.HandleFunc("/{city}/restaurants/recommended",
                controller.NewTopRatedHandler(controller.Dependencies{
                    GetTopRestaurants: topRatedStub,
                }),
            )

            req := httptest.NewRequest(
                http.MethodGet,
                "/vancouverbc/restaurants/recommended",
                nil,
            )
            w := httptest.NewRecorder()
            mux.ServeHTTP(w, req)

            res := w.Result()
            if res.StatusCode != http.StatusOK {
                t.Fatalf("expected 200, got %d", res.StatusCode)
            }
            ct := res.Header.Get("Content-Type")
            if !strings.Contains(strings.ToLower(ct), "json") {
                t.Fatalf("expected json content-type, got %q", ct)
            }

            var payload struct {
                Restaurants []struct {
                    ID string `json:"id"`
                } `json:"restaurants"`
            }
            if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
                t.Fatalf("decode: %v", err)
            }
            if len(payload.Restaurants) != 2 {
                t.Fatalf("expected 2, got %d", len(payload.Restaurants))
            }
            if payload.Restaurants[0].ID != "r1" || payload.Restaurants[1].ID != "r2" {
                t.Errorf("unexpected restaurants: %v", payload.Restaurants)
            }
        })
    }

In Go, there is no need for a `replaceFactoriesForTest` mechanism. Because the wiring is explicit Go code (not a DI framework's reflection-based registry), you simply construct the handler with whatever dependencies you want in tests. There's no "production factory" that needs to be overridden; the test constructs everything directly. This is one place where Go's explicitness pays off.

Next, I write the code that assembles my modules.

cmd/server/main.go…

    func main() {
        db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
        if err != nil {
            log.Fatalf("open db: %v", err)
        }

        // Wire dependencies bottom-up, explicitly.
        // Nothing here requires reflection or a DI framework.
        topRatedDeps := toprated.Dependencies{
            FindRatingsByRestaurant: func(ctx context.Context, city string) ([]toprated.RatingsByRestaurant, error) {
                panic("not yet implemented")    ➀
            },
            CalculateRatingForRestaurant: func(r toprated.RatingsByRestaurant) int {
                panic("not yet implemented")
            },
        }
        getTopRestaurants := toprated.New(topRatedDeps)

        handlerDeps := controller.Dependencies{
            GetTopRestaurants: getTopRestaurants,
        }
        handler := controller.NewTopRatedHandler(handlerDeps)

        mux := http.NewServeMux()
        mux.HandleFunc("/{city}/restaurants/recommended", handler)

        log.Fatal(http.ListenAndServe(":3000", mux))
    }

controller.go toprated.go

handler()

topRated()

main.go

GetTopRestaurants() FindRatingsByRestaurant() CalculateRatings
ForRestaurants()

Sometimes I have a dependency for a module defined but nothing to fulfill that contract yet. That is totally fine. I can just define an implementation inline that panics as in the `topRatedDeps` above. Acceptance tests will fail but, at this stage, that is as I would expect.

Finding and fixing a problem
----------------------------

The careful observer will notice that there is a compile error at the point the handler is constructed because I have a conflict between two definitions:

*   the representation of the restaurant as understood by `controller.go`
*   the restaurant as defined in `toprated.go` and returned by `GetTopRestaurants`.

The reason is simple: I have yet to add a `Name` field to the `Restaurant` type in `toprated.go`. There is a trade-off here. If I had a single type representing a restaurant, rather than one in each module, I would only have to add `Name` once, and both modules would compile without additional changes. Nonetheless, I choose to keep the types separate, even though it creates extra template code. By maintaining two distinct types, one for each layer of my application, I'm much less likely to couple those layers unnecessarily. No, this is not very [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself), but I am often willing to risk some repetition to keep the module contracts as independent as possible.

src/restaurantratings/toprated.go…

    type Restaurant struct {
        ID   string
        Name string
    }

    func toRestaurant(r overallRating) Restaurant {
        return Restaurant{
            ID:   r.restaurantID,
            Name: "",    // TODO: dummy value to make the contract compile; add mapping test shortly
        }
    }

My extremely naive solution gets the code compiling again, allowing me to continue on my current work on the module. I'll shortly add validation to my tests that ensure that the `Name` field is mapped as it should be. Now with the test passing, I move on to the next step, which is to provide a more permanent solution to the restaurant mapping.

Method values in Go
-------------------

In Go, there is no `this` problem. When you pass a method as a function value, Go binds the receiver at the point you take the value:

    repo := &RestaurantRepository{db: db}
    getByID := repo.FindByID    // repo is bound; getByID is a plain func(ctx, id) (Restaurant, error)

The result is an ordinary function value. There is no risk of a lost receiver; Go's method values capture `repo` by pointer, as expected. You can pass `getByID` directly as a `func(context.Context, string) (Restaurant, error)` dependency without any wrapper. Unlike JavaScript where `this` depends on how a function is called — requiring `.bind()` or arrow function wrappers to stabilize it — Go's method value syntax is simply a bound function reference. No wrappers, no adapters.

Reaching out to the repository layer
-------------------------------------

Now, with the structure of my `GetTopRestaurants` function more or less in place and in need of a way to get the restaurant name, I will fill out the restaurant loading to fetch the rest of the `Restaurant` data. In the past, before adopting this highly function-driven style of development, I probably would have built a repository interface or struct with a method meant to load the `Restaurant` object. Now my inclination is to build the minimum I need: a function definition for loading the object without making any assumptions about the implementation. That can come later when I'm binding to that function.

test/restaurantratings/toprated\_test.go…

    restaurantsById := map[string]toprated.Restaurant{
        "restaurant1": {ID: "restaurant1", Name: "Restaurant 1"},
        "restaurant2": {ID: "restaurant2", Name: "Restaurant 2"},
    }

    getRestaurantByIDStub := func(ctx context.Context, id string) (toprated.Restaurant, error) {    ➀
        r, ok := restaurantsById[id]
        if !ok {
            return toprated.Restaurant{}, fmt.Errorf("not found: %s", id)
        }
        return r, nil
    }

    // SNIP...

    deps := toprated.Dependencies{
        GetRestaurantByID:            getRestaurantByIDStub,    ➁
        FindRatingsByRestaurant:      findRatingsByRestaurantStub,
        CalculateRatingForRestaurant: calculateRatingForRestaurantStub,
    }

    getTopRated := toprated.New(deps)
    topRestaurants, err := getTopRated(context.Background(), "vancouverbc")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if topRestaurants[0].ID != "restaurant1" || topRestaurants[0].Name != "Restaurant 1" {    ➂
        t.Errorf("unexpected first restaurant: %+v", topRestaurants[0])
    }
    if topRestaurants[1].ID != "restaurant2" || topRestaurants[1].Name != "Restaurant 2" {
        t.Errorf("unexpected second restaurant: %+v", topRestaurants[1])
    }

In my domain-level test, I have introduced:

1.  a stubbed finder for the `Restaurant` — signature `func(ctx context.Context, id string) (Restaurant, error)`.
2.  an entry in my dependencies for that finder.
3.  validation that the name matches what was loaded from the `Restaurant` object.

As with previous functions that load data, `GetRestaurantByID` returns `(Restaurant, error)`. Although I continue to play the little game, pretending that I don't know how I will implement the function, I know the `Restaurant` is coming from an external data source. This makes the mapping code more involved.

src/restaurantratings/toprated.go…

    func (deps Dependencies) getTopRestaurants(ctx context.Context, city string) ([]Restaurant, error) {
        ratingsByRestaurant, err := deps.FindRatingsByRestaurant(ctx, city)
        if err != nil {
            return nil, fmt.Errorf("find ratings: %w", err)
        }

        overallRatings := calculateRatings(ratingsByRestaurant, deps.CalculateRatingForRestaurant)
        sortByOverallRating(overallRatings)

        return loadRestaurants(ctx, overallRatings, deps.GetRestaurantByID)    ➀
    }

    func loadRestaurants(
        ctx context.Context,
        ratings []overallRating,
        getByID func(context.Context, string) (Restaurant, error),
    ) ([]Restaurant, error) {
        restaurants := make([]Restaurant, len(ratings))
        g, ctx := errgroup.WithContext(ctx)    ➁
        for i, r := range ratings {
            i, r := i, r    // capture loop variables
            g.Go(func() error {
                restaurant, err := getByID(ctx, r.restaurantID)
                if err != nil {
                    return fmt.Errorf("load restaurant %s: %w", r.restaurantID, err)
                }
                restaurants[i] = restaurant
                return nil
            })
        }
        if err := g.Wait(); err != nil {
            return nil, err
        }
        return restaurants, nil
    }

1.  The complexity comes from the fact that `loadRestaurants` fires goroutines for each lookup.
2.  `golang.org/x/sync/errgroup` is the Go equivalent of `Promise.all`: it fans out work across goroutines, waits for all to complete, and returns the first error (if any). Unlike `Promise.all`, it also propagates cancellation via the derived `ctx`, so a single failed lookup can cancel the remaining ones.

I don't want each of these requests to block, or my IO-bound loads will run serially, delaying the entire user request, but I need to block until all the lookups are complete. `errgroup` handles both. This is fine for a top 10 list since the number of concurrent goroutines is small. In an application of any scale, I would probably restructure my service calls to load the `Name` field via a database join and eliminate the extra call. If that option was not available, for example, I was querying an external API, I might wish to bound concurrency using a semaphore or `errgroup` with a limit (available in newer versions of `golang.org/x/sync`).

Again, I update my assembly module with a dummy implementation so it all compiles, then start on the code that fulfills my remaining contracts.

cmd/server/main.go…

    topRatedDeps := toprated.Dependencies{
        FindRatingsByRestaurant: func(ctx context.Context, city string) ([]toprated.RatingsByRestaurant, error) {
            panic("not yet implemented")
        },
        CalculateRatingForRestaurant: func(r toprated.RatingsByRestaurant) int {
            panic("not yet implemented")
        },
        GetRestaurantByID: func(ctx context.Context, id string) (toprated.Restaurant, error) {
            panic("not yet implemented")
        },
    }

controller.go toprated.go

handler()

topRated()

main.go

GetTopRestaurants() FindRatingsByRestaurant() CalculateRatings
ForRestaurants() GetRestaurantByID()

The last mile: implementing domain layer dependencies
-----------------------------------------------------

With my handler and main domain module workflow in place, it's time to implement the dependencies, namely the database access layer and the weighted rating algorithm.

This leads to the following set of high-level functions and dependencies

controller.go toprated.go ratingsalgorithm.go restaurantrepo.go ratingsrepo.go

handler()

topRated()

main.go

CalculateRatings
ForRestaurants()

groupedBy
Restaurant()

findByID()

GetTopRestaurants() FindRatingsByRestaurant() CalculateRatings
ForRestaurants() GetRestaurantByID()

For testing, I have the following arrangement of stubs

controller.go toprated.go

handler()

topRated()

calculateRatingFor
RestaurantStub()

findRatingsBy
RestaurantStub

getRestaurantBy
IDStub()

GetTopRestaurants() FindRatingsByRestaurant() CalculateRatings
ForRestaurants() GetRestaurantByID()

For testing, all the elements are created by the test code, but I haven't shown that in the diagram due to clutter.

The process for implementing these modules follows the same pattern:

*   implement a test to drive out the basic design and a `Dependencies` struct if one is necessary
*   build the basic logical flow of the module, making the test pass
*   implement the module dependencies
*   repeat.

I won't walk through the entire process again since I've already demonstrated the process. The code for the modules working end-to-end is available [in the repo](https://github.com/danielsomerfield/function-first-composition-example-go). Some aspects of the final implementation require additional commentary.

By now, you might expect my ratings algorithm to be made available via yet another factory implemented as a closure over a dependencies struct. This time I chose to write a pure function instead.

src/restaurantratings/ratingsalgorithm.go…

    type RestaurantRating struct {
        Rating      Rating
        RatedByUser User
    }

    type User struct {
        ID        string
        IsTrusted bool
    }

    type RatingsByRestaurant struct {
        RestaurantID string
        Ratings      []RestaurantRating
    }

    func CalculateForRestaurant(ratings RatingsByRestaurant) int {
        total := 0
        for _, r := range ratings.Ratings {
            multiplier := 1
            if r.RatedByUser.IsTrusted {
                multiplier = 4
            }
            total += int(r.Rating) * multiplier
        }
        return total
    }

I made this choice to signal that this should always be a simple, stateless calculation. Had I wanted to leave an easy pathway toward a more complex implementation, say something backed by a data science model parameterized per user, I would have used the factory pattern again. Often there isn't a right or wrong answer. The design choice provides a trail, so to speak, indicating how I anticipate the software might evolve. I create more rigid code in areas that I don't think should change while leaving more flexibility in the areas I have less confidence in the direction.

Another example where I "leave a trail" is the decision to define another `RestaurantRating` type in `ratingsalgorithm.go`. The type is structurally identical to `RestaurantRating` defined in `toprated.go`. I could take another path here:

*   export `RestaurantRating` from `toprated` and reference it directly in `ratingsalgorithm`, or
*   factor `RestaurantRating` out into a common package. You will often see shared definitions in a package called `types` or `models`, although I prefer a more contextual name like `domain` which gives some hints about the kind of types contained therein.

In this case, I am not confident that these types are literally the same. They might be different projections of the same domain entity with different fields, and I don't want to share them across the package boundaries risking deeper coupling. As unintuitive as this may seem, I believe it is the right choice: collapsing the entities is very cheap and easy at this point. If they begin to diverge, I probably shouldn't merge them anyway, but pulling them apart once they are bound can be very tricky.

If it looks like a duck
-----------------------

I promised to explain why I often choose not to export types. I want to make a type available to another package only if I am confident that doing so won't create incidental coupling, restricting the ability of the code to evolve.

Go has a dual nature here that is worth understanding precisely:

*   **Go interfaces are structural** — any type that has the right methods satisfies an interface without declaring so. As long as two types agree on a method set, they are compatible. This is Go's "duck typing": if it looks like a duck and quacks like a duck, it is a duck, for interface-satisfaction purposes.
*   **Go structs are nominal** — a struct in package `toprated` and an identical struct in package `ratingsalgorithm` are distinct types. You cannot pass one where the other is expected. The compiler does not collapse them.

This dual nature matters for how we approach the module-decoupling strategy. For function-typed fields in `Dependencies` structs, the structural nature of function types means that any function with the right signature satisfies the field — regardless of where it was defined. The stub in the test and the real implementation in the repository both satisfy `func(context.Context, string) (Restaurant, error)` if their signatures match. No shared interface declaration required.

When you need a contract that a caller can hold without importing the callee's concrete type, you have four options:

1.  **Local interface**: Define a narrow interface in the _caller's_ package, containing only the methods the caller needs. Any type from any package that has those methods satisfies it, without the callee knowing the interface exists. This is the preferred Go idiom when the consumer knows what it needs but shouldn't dictate the provider's full shape.
2.  **Shared package**: Factor the shared type into a common package (e.g., `domain`). Both caller and callee import from there. This is appropriate when the type is genuinely universal — a first-class domain concept that appears everywhere, and the coupling is intentional. Accept that changing the shared type will require updating all consumers.
3.  **Duplicate types + adapter**: Define the type separately in each package and write an adapter function at the seam. This is verbose but keeps the packages fully independent. Useful when the types will likely diverge over time.
4.  **Function-typed fields**: Rather than a method interface, express the dependency as a `func(...)` field in a `Dependencies` struct. No interface declaration needed anywhere. This is the approach this article has been demonstrating throughout, and it works particularly well for narrow, single-operation dependencies.

Option 4 is almost free of ceremony: no interface declaration, no mock generation, just a function literal in the test and a real function value in `main.go`. It scales down gracefully — a dependency on a single operation does not need an interface with a single method; it can just be a field of function type.

The local interface (option 1) is the right complement when you do need to express a multi-method contract or when you are consuming a third-party type and want to insulate yourself from it. Define the interface in the package that needs the contract, not in the package that implements it.

The shared `domain` package (option 2) remains pragmatic for types that are genuinely universal domain entities — a `RestaurantID`, a `Rating`, a `UserID` — things that appear in acceptance criteria by name and are unlikely to split into diverging projections. The test for "should this be shared?" is: if two packages have structurally identical types and you are confident they will always be structurally identical, the duplicate is pure noise. If you are not confident, keep them separate.

For the Go port of this project, see [github.com/danielsomerfield/function-first-composition-example-go](https://github.com/danielsomerfield/function-first-composition-example-go).

In summary
----------

By choosing to fulfill dependency contracts with function-typed fields in structs rather than with classes or interface registries, minimizing the code sharing between packages and driving the design through tests, I can create a Go system composed of highly discrete, evolvable, but still type-safe modules. The approach maps naturally onto Go idioms:

*   `Dependencies` structs with function fields replace interface-based injection without requiring interface declarations.
*   `context.Context` as the first parameter of every IO-bound function makes cancellation and deadline propagation explicit throughout.
*   `_test.go` scoping enforces the "don't export until certain" discipline mechanically.
*   `net/http/httptest` makes handler testing first-class without custom stubs.
*   `errgroup` provides composable parallel IO with error propagation and cancellation.
*   Method values eliminate the `this`-binding problem entirely.

If you have similar priorities in your next Go project, consider adopting some aspects of the approach I have outlined. Be aware, however, that choosing a foundational approach for your project is rarely as simple as selecting the "best practice" — it requires taking into account other factors, such as the idioms of your tech stack and the skills of your team. There are many ways to put a system together, each with a complex set of tradeoffs. That makes software architecture often difficult and always engaging. I wouldn't have it any other way.

* * *

Significant Revisions

_25 March 2026:_ Adapted for Go (closures over dependency structs, `context.Context`, `errgroup`, `httptest`, `_test.go` scoping)

_03 July 2023:_ Add references to Go and Kotlin ports of example project

_23 May 2023:_ Published
