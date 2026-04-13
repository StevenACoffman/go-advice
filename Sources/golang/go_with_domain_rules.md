# Go with the Domain — Rules for Claude

Distilled from *Go with the Domain* by Robert Laszczak and Miłosz Smółka (Three Dots Labs).
Covers chapters 6–17 of the book (Wild Workouts example application).

These rules apply to business-oriented Go services — API servers, SaaS backends, anything with domain logic worth protecting. Skip them for simple CRUD utilities.

---

## 1. When NOT to Use DRY — The Shared Models Problem

**Never use the same struct for HTTP responses, database storage, and domain logic.**

Each layer has different change drivers. When they share a struct, a database column change silently breaks the API, or an API change corrupts stored data.

The canonical example from the book: a `User` struct served as both a Firestore document model and an OpenAPI HTTP response. Susan, a new engineer, needed to add a `LastIP` field for database-only storage. She added it to the shared struct and then had to nil it out in the HTTP handler with a comment — a fragile, silent coupling that anyone on the team could miss:

```go
// BAD: one struct serves everything
type User struct {
    Balance     int    `json:"balance"`
    DisplayName string `json:"displayName"`
    Role        string `json:"role"`
    LastIp      string `json:"lastIp"` // should NOT appear in API response
}

// HTTP handler had to manually suppress the field:
user.LastIp = nil
render.Respond(w, r, user)
```

The fix: separate models per layer.

```go
// database model (adapters layer) — firestore.go
type UserModel struct {
    Balance     int
    DisplayName string
    Role        string
    LastIP      string
}

// HTTP response (ports layer / OpenAPI generated) — http.go
userResponse := User{
    DisplayName: authUser.DisplayName,
    Balance:     user.Balance,
    Role:        authUser.Role,
}
render.Respond(w, r, userResponse)
```

**Rule:** DRY applies to *behaviour* (shared functions), not to *data* (shared structs). Code duplication between layers is intentional and good.

The question to ask: "Is this data coupling between concerns that should be able to evolve independently?" If yes, split the structs. If the code using the common structure is likely to change together, a shared struct is fine.

**Microservices don't fix this.** Starting with poorly separated services leads to a distributed monolith — all the coupling of a monolith plus network overhead. Splitting by entity or database table without domain boundaries creates high coupling between services.

---

## 2. Domain-Driven Design Lite

### 2a. Three explicit DDD Lite rules — stated in this exact order

The book states three DDD Lite rules verbatim:

1. **Reflect business logic literally** — domain code should read like the business talks. The test: "Will a non-technical person understand my code without any extra translation of technical terms?" When you talk with business stakeholders, they say "I'm scheduling training on 13:00", not "I'm setting the attribute state to 'training_scheduled' for hour 13:00."

2. **Always keep a valid state in memory** — domain objects must never be in an invalid or partially-constructed state. Private fields, constructors that validate, and methods that enforce invariants. The Rugged Manifesto: "I recognize that my code will be used in ways I cannot anticipate."

3. **Domain must be database-agnostic** — no database tags, ORM annotations, or infrastructure imports on domain types. Domain types are shaped only by business rules.

**Tactical vs Strategic split:** Tactical DDD patterns alone deliver roughly 30% of DDD's value. The remaining 70% comes from Strategic patterns: Event Storming, bounded contexts, and ubiquitous language built with domain experts. Many teams use only tactical patterns without the strategic part — the book calls this "super horrifying."

### 2b. Model behaviour, not data

Domain types must expose *behaviours*, not getters/setters. The method name should match the ubiquitous language your business uses.

The original Wild Workouts gRPC handler had 50+ lines checking boolean flags and setting them directly. After DDD Lite refactoring, the handler dropped to 18 lines with no domain logic:

```go
// BAD: data-oriented — flag checking and direct mutation in handler
func (g GrpcServer) UpdateHour(ctx context.Context, req *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
    // ...
    if req.HasTrainingScheduled && !hour.Available {
        return nil, status.Error(codes.FailedPrecondition, "hour is not available for training")
    }
    if req.Available && req.HasTrainingScheduled {
        return nil, status.Error(codes.FailedPrecondition, "cannot set hour as available when it have training scheduled")
    }
    if !req.Available && !req.HasTrainingScheduled {
        return nil, status.Error(codes.FailedPrecondition, "cannot set hour as unavailable when it have no training scheduled")
    }
    hour.Available = req.Available
    // ... 30+ more lines
}
```

```go
// GOOD: 18-line handler, all domain logic in the domain type
func (g GrpcServer) MakeHourAvailable(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
    trainingTime, err := protoTimestampToTime(request.Time)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "unable to parse time")
    }

    if err := g.hourRepository.UpdateHour(ctx, trainingTime, func(h *hour.Hour) (*hour.Hour, error) {
        if err := h.MakeAvailable(); err != nil {
            return nil, err
        }
        return h, nil
    }); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &trainer.EmptyResponse{}, nil
}
```

The gRPC API itself was also changed from one CRUD method to three behaviour-oriented methods:

```protobuf
// BAD: one god method with boolean flags
rpc UpdateHour(UpdateHourRequest) returns (EmptyResponse) {}
message UpdateHourRequest {
  google.protobuf.Timestamp time = 1;
  bool has_training_scheduled = 2;
  bool available = 3;
}

// GOOD: separate methods matching business behaviours
rpc ScheduleTraining(UpdateHourRequest) returns (EmptyResponse) {}
rpc CancelTraining(UpdateHourRequest) returns (EmptyResponse) {}
rpc MakeHourAvailable(UpdateHourRequest) returns (EmptyResponse) {}
message UpdateHourRequest {
  google.protobuf.Timestamp time = 1;
  // no flags needed — behaviour is in the method name
}
```

### 2c. The `Hour` type — private fields, valid state, behaviour methods

```go
type Hour struct {
    hour         time.Time
    availability Availability // private — callers cannot corrupt state
}

// Availability is a value object — a type with a constrained set of valid states
type Availability int

const (
    Available         Availability = iota
    NotAvailable      Availability = iota
    TrainingScheduled Availability = iota
)
```

All methods enforce invariants. The caller cannot set state directly:

```go
func (h *Hour) ScheduleTraining() error {
    if !h.IsAvailable() {
        return ErrHourNotAvailable
    }
    h.availability = TrainingScheduled
    return nil
}

func (h *Hour) CancelTraining() error {
    if !h.HasTrainingScheduled() {
        return ErrNoTrainingScheduled
    }
    h.availability = Available
    return nil
}

func (h *Hour) MakeAvailable() error {
    if h.HasTrainingScheduled() {
        return errors.New("cannot make hour with training scheduled available")
    }
    h.availability = Available
    return nil
}

func (h *Hour) MakeNotAvailable() error {
    if h.HasTrainingScheduled() {
        return errors.New("cannot make hour with training scheduled not available")
    }
    h.availability = NotAvailable
    return nil
}
```

Bad pattern — caller controls state transitions:

```go
h := hour.NewAvailableHour("13:00")
if h.HasTrainingScheduled() {
    h.SetState(hour.Available)
} else {
    return errors.New("unable to cancel training")
}
```

Good pattern — domain type controls its own transitions:

```go
if err := h.CancelTraining(); err != nil {
    return err
}
```

Validation lives in exactly one place: the constructor and mutating methods of the domain type. No dumb getters or setters.

### 2d. `NewAvailableHour` vs `UnmarshalHourFromDatabase` — two constructors, different purposes

**Why two constructors?** When loading from a database, you need to reconstruct a domain object *without running the normal business-rule validation*. For example, an archived hour may be in a state that would be rejected by `NewAvailableHour`, but is perfectly valid stored data.

```go
// NewAvailableHour validates business rules — use when creating new objects
func NewAvailableHour(hour time.Time) (*Hour, error) {
    if err := validateTime(hour); err != nil {
        return nil, err
    }
    return &Hour{
        hour:         hour,
        availability: Available,
    }, nil
}

// UnmarshalHourFromDatabase skips business-rule validation — use only in adapters layer
// Naming convention: Unmarshal*FromDatabase
func UnmarshalHourFromDatabase(hour time.Time, availability Availability) (*Hour, error) {
    if err := validateTime(hour); err != nil { // basic sanity only, not business rules
        return nil, err
    }
    return &Hour{hour: hour, availability: availability}, nil
}
```

`validateTime` checks basic sanity (non-zero time, within valid range) but does NOT enforce business rules like "must be in the future." Stored data may legitimately violate business rules that apply at creation time.

### 2e. The `Factory` type for reconstruction

The `Factory` type holds configuration and provides both constructors. This keeps configuration (like `MaxWeeksInTheFutureToSet`, `MinUtcHour`, `MaxUtcHour`) in one place:

```go
type Factory struct {
    fc FactoryConfig
}

type FactoryConfig struct {
    MaxWeeksInTheFutureToSet int
    MinUtcHour               int
    MaxUtcHour               int
}

func (f Factory) NewAvailableHour(hour time.Time) (*Hour, error) {
    // validates using f.fc config
}

func (f Factory) UnmarshalHourFromDatabase(hour time.Time, availability Availability) (*Hour, error) {
    // skips business-rule validation, uses basic sanity check only
}
```

In the MySQL repository:

```go
domainHour, err := m.hourFactory.UnmarshalHourFromDatabase(dbHour.Hour.Local(), availability)
```

In the Firestore repository:

```go
return f.hourFactory.UnmarshalHourFromDatabase(firebaseHour.Hour.Local(), availability)
```

**Never call `NewAvailableHour` inside the repository** — it's the wrong constructor for reconstruction. The factory separates "creating a new valid domain object" from "restoring a stored domain object."

### 2f. The `Training` entity

```go
package training

type Training struct {
    uuid            string
    userUUID        string
    userName        string
    time            time.Time
    notes           string
    proposedNewTime time.Time
    moveProposedBy  UserType
    canceled        bool
}

func NewTraining(uuid string, userUUID string, userName string, trainingTime time.Time) (*Training, error) {
    if uuid == "" {
        return nil, errors.New("empty training uuid")
    }
    if userUUID == "" {
        return nil, errors.New("empty userUUID")
    }
    if userName == "" {
        return nil, errors.New("empty userName")
    }
    if trainingTime.IsZero() {
        return nil, errors.New("zero training time")
    }
    return &Training{
        uuid:     uuid,
        userUUID: userUUID,
        userName: userName,
        time:     trainingTime,
    }, nil
}
```

All fields private. Constructor validates all required fields. Methods enforce business rules.

Key methods on `Training`:

```go
func (t *Training) Cancel() error {
    if t.IsCanceled() {
        return ErrTrainingAlreadyCanceled
    }
    t.canceled = true
    return nil
}

func (t Training) CanBeCanceledForFree() bool {
    return t.time.Sub(time.Now()) >= time.Hour*24
}

func (t *Training) ApproveReschedule(userType UserType) error {
    if !t.IsRescheduleProposed() {
        return errors.WithStack(ErrNoRescheduleRequested)
    }
    if t.moveProposedBy == userType {
        return errors.Errorf(
            "trying to approve reschedule by the same user type which proposed reschedule (%s)",
            userType.String(),
        )
    }
    t.time = t.proposedNewTime
    t.proposedNewTime = time.Time{}
    t.moveProposedBy = UserType{}
    return nil
}
```

### 2g. Pure domain functions for business calculations

Not everything needs to be a method. When a calculation is a pure function of domain values, write it as a package-level function. In Go, this is idiomatic — don't force everything into methods on a type.

`CancelBalanceDelta` is the canonical example. It encodes a business rule (the training balance to adjust after cancellation) without being a method, because it doesn't need to mutate state:

```go
// CancelBalanceDelta returns the training balance delta that should be adjusted
// after training cancellation. Domain rule:
// - cancelling >= 24h before: attendee gets training back (+1)
// - trainer cancels < 24h before: attendee gets training back + 1 fine (+2)
// - attendee cancels < 24h before: attendee loses training, no fine (0)
func CancelBalanceDelta(tr Training, cancelingUserType UserType) int {
    if tr.CanBeCanceledForFree() {
        // just give training back
        return 1
    }

    switch cancelingUserType {
    case Trainer:
        // 1 for cancelled training +1 "fine" for cancelling by trainer less than 24h before training
        return 2
    case Attendee:
        // "fine" for cancelling less than 24h before training
        return 0
    default:
        panic(fmt.Sprintf("not supported user type %s", cancelingUserType))
    }
}
```

The book notes: "If you start to see `if`s related to logic in your application layer, you should think about how to move it to the domain layer." `CancelBalanceDelta` moved from the application service into the domain package.

### 2h. Database-agnostic domain

No database tags, ORM annotations, or `db:""` struct tags on domain types. The DB type lives only in the adapters layer:

```go
// BAD: domain type polluted with DB concerns
type Hour struct {
    Hour         time.Time `db:"hour"`
    Availability string    `db:"availability"`
}

// GOOD: DB type is separate, in adapters package only
type mysqlHour struct {
    ID           string    `db:"id"`
    Hour         time.Time `db:"hour"`
    Availability string    `db:"availability"`
}
```

Go's design with no "magic" like ORM annotations means database solutions affect code in an even more significant way than in other languages. This makes the separation even more important.

### 2i. Domain-First approach

Start implementing with an in-memory repository. Get the domain right using only unit tests before choosing a database. With complex projects, you can spend 2–4 weeks working only on the domain layer.

Deferring the database decision gives you more information and context to make a better choice. This approach requires a good relationship and trust from the business. Strategic DDD patterns (Event Storming, bounded contexts) will help build that trust.

### 2j. Must-constructors in tests

Provide `Must*` constructors in tests to reduce boilerplate. They panic on invalid input — acceptable in test code:

```go
func NewUser(userUUID string, userType UserType) (User, error) {
    if userUUID == "" {
        return User{}, errors.New("missing user UUID")
    }
    if userType.IsZero() {
        return User{}, errors.New("missing user type")
    }
    return User{userUUID: userUUID, userType: userType}, nil
}

func MustNewUser(userUUID string, userType UserType) User {
    u, err := NewUser(userUUID, userType)
    if err != nil {
        panic(err)
    }
    return u
}
```

Similarly provide `newExampleTrainingWithTime`, `newCanceledTraining`, and similar helpers to make domain tests readable.

---

## 3. The Repository Pattern

### 3a. Define the interface in the domain package

```go
package hour

type Repository interface {
    GetOrCreateHour(ctx context.Context, hourTime time.Time) (*Hour, error)
    UpdateHour(
        ctx context.Context,
        hourTime time.Time,
        updateFn func(h *Hour) (*Hour, error),
    ) error
}
```

The interface lives next to the type it operates on — same package as `Hour`. Implementations live in `adapters/`. This follows the same pattern as `io.Writer`, where the `io` package defines the interface and all implementations are decoupled in separate packages.

For the `Training` aggregate (more complex), the interface includes `user` as an explicit parameter for secure-by-design access control (see section 6):

```go
package training

type Repository interface {
    AddTraining(ctx context.Context, tr *Training) error
    GetTraining(ctx context.Context, trainingUUID string, user User) (*Training, error)
    UpdateTraining(
        ctx context.Context,
        trainingUUID string,
        user User,
        updateFn func(ctx context.Context, tr *Training) (*Training, error),
    ) error
}
```

The original Wild Workouts repository had separate methods for each use case:

```go
// BAD: before DDD refactoring — one method per use case
type trainingRepository interface {
    FindTrainingsForUser(ctx context.Context, user auth.User) ([]Training, error)
    AllTrainings(ctx context.Context) ([]Training, error)
    CreateTraining(ctx context.Context, training Training, createFn func() error) error
    CancelTraining(ctx context.Context, trainingUUID string, deleteFn func(Training) error) error
    RescheduleTraining(ctx context.Context, trainingUUID string, newTime time.Time, updateFn func(Training) (Training, error)) error
    ApproveTrainingReschedule(ctx context.Context, trainingUUID string, updateFn func(Training) (Training, error)) error
    RejectTrainingReschedule(ctx context.Context, trainingUUID string, updateFn func(Training) (Training, error)) error
}
```

After introducing `training.Training`, it collapsed to three methods. Repositories are "stupid" — they save and load domain objects, not validate domain rules.

### 3b. Use the closure pattern for transactions — the `updateFn` pattern

Pass an `updateFn` closure into the update method. The repository handles the transaction; the caller handles the business logic. This is the author's preferred approach after experimenting with passing transactions via `context.Context` and handling transactions at HTTP/gRPC middleware level, both of which had issues with being "magical, not explicit, and slow in some cases."

The pattern in one transaction:
1. Get/create the entity (with `FOR UPDATE` for SQL to prevent parallel updates)
2. Execute the closure (caller's business logic)
3. Save the returned entity
4. Rollback on any error returned from the closure

The MySQL implementation:

```go
func (m MySQLHourRepository) UpdateHour(
    ctx context.Context,
    hourTime time.Time,
    updateFn func(h *hour.Hour) (*hour.Hour, error),
) (err error) {
    tx, err := m.db.Beginx()
    if err != nil {
        return errors.Wrap(err, "unable to start transaction")
    }

    // Named return (err error) lets defer see the final return value.
    // Even if function exits without error, commit still can return error.
    // In that case we can override nil to err with `err = m.finishTransaction(...)`.
    defer func() {
        err = m.finishTransaction(err, tx)
    }()

    existingHour, err := m.getOrCreateHour(ctx, tx, hourTime, true) // true = FOR UPDATE
    if err != nil {
        return err
    }

    updatedHour, err := updateFn(existingHour)
    if err != nil {
        return err
    }

    if err := m.upsertHour(tx, updatedHour); err != nil {
        return err
    }

    return nil
}
```

### 3c. `finishTransaction` with named returns and `multierr`

```go
// finishTransaction rollbacks transaction if error is provided.
// If err is nil transaction is committed.
// If the rollback fails, we are using multierr library to add error about rollback failure.
// If the commit fails, commit error is returned.
func (m MySQLHourRepository) finishTransaction(err error, tx *sqlx.Tx) error {
    if err != nil {
        if rollbackErr := tx.Rollback(); rollbackErr != nil {
            return multierr.Combine(err, rollbackErr)
        }
        return err
    } else {
        if commitErr := tx.Commit(); commitErr != nil {
            return errors.Wrap(err, "failed to commit tx")
        }
        return nil
    }
}
```

The `(err error)` named return in `UpdateHour` is critical: the deferred `finishTransaction` can see whether the function is exiting with an error, and can also override a nil return to an error if the commit fails.

### 3d. `sqlContextGetter` interface — abstracting tx vs db

Define a local interface that both `*sqlx.Tx` and `*sqlx.DB` satisfy. This lets `getOrCreateHour` work in both transactional and non-transactional contexts without duplication:

```go
// sqlContextGetter is an interface provided both by transaction and standard db connection
type sqlContextGetter interface {
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (m MySQLHourRepository) GetOrCreateHour(ctx context.Context, time time.Time) (*hour.Hour, error) {
    return m.getOrCreateHour(ctx, m.db, time, false)
}

func (m MySQLHourRepository) getOrCreateHour(
    ctx context.Context,
    db sqlContextGetter,
    hourTime time.Time,
    forUpdate bool,
) (*hour.Hour, error) {
    dbHour := mysqlHour{}

    query := "SELECT * FROM `hours` WHERE `hour` = ?"
    if forUpdate {
        query += " FOR UPDATE"
    }

    err := db.GetContext(ctx, &dbHour, query, hourTime.UTC())
    if errors.Is(err, sql.ErrNoRows) {
        // in reality this date exists, even if it's not persisted
        return m.hourFactory.NewNotAvailableHour(hourTime)
    } else if err != nil {
        return nil, errors.Wrap(err, "unable to get hour from db")
    }

    availability, err := hour.NewAvailabilityFromString(dbHour.Availability)
    if err != nil {
        return nil, err
    }

    domainHour, err := m.hourFactory.UnmarshalHourFromDatabase(dbHour.Hour.Local(), availability)
    if err != nil {
        return nil, err
    }

    return domainHour, nil
}
```

`FOR UPDATE` blocks other transactions from modifying the same rows until the current transaction commits or rolls back. Without it, two parallel transactions can both read "available" and both schedule a training for the same hour.

### 3e. In-memory repository stores values, not pointers

```go
type MemoryHourRepository struct {
    hours map[time.Time]hour.Hour // values, not *hour.Hour
    lock  *sync.RWMutex

    hourFactory hour.Factory
}

func NewMemoryHourRepository(hourFactory hour.Factory) *MemoryHourRepository {
    if hourFactory.IsZero() {
        panic("missing hourFactory")
    }
    return &MemoryHourRepository{
        hours:       map[time.Time]hour.Hour{},
        lock:        &sync.RWMutex{},
        hourFactory: hourFactory,
    }
}

func (m MemoryHourRepository) getOrCreateHour(hourTime time.Time) (*hour.Hour, error) {
    currentHour, ok := m.hours[hourTime]
    if !ok {
        return m.hourFactory.NewNotAvailableHour(hourTime)
    }
    // we don't store hours as pointers, but as values
    // thanks to that, we are sure that nobody can modify Hour without using UpdateHour
    return &currentHour, nil
}

func (m *MemoryHourRepository) UpdateHour(
    _ context.Context,
    hourTime time.Time,
    updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
    m.lock.Lock()
    defer m.lock.Unlock()

    currentHour, err := m.getOrCreateHour(hourTime)
    if err != nil {
        return err
    }

    updatedHour, err := updateFn(currentHour)
    if err != nil {
        return err
    }

    m.hours[hourTime] = *updatedHour // store value — the "commit"
    return nil
}
```

If you store pointers, callers can mutate the stored object without going through `UpdateHour`, bypassing all validation. Storing values forces callers to go through the proper update path. The "commit" is `m.hours[hourTime] = *updatedHour` — if `updateFn` returns an error and you return early, no commit happens.

### 3f. Firestore transactions — `getDocumentFn` closure

Firebase's transactional and non-transactional query interfaces are not fully compatible. Use a `getDocumentFn` closure to avoid duplicating the document-fetch code:

```go
func (f FirestoreHourRepository) UpdateHour(
    ctx context.Context,
    hourTime time.Time,
    updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
    err := f.firestoreClient.RunTransaction(ctx, func(ctx context.Context, transaction *firestore.Transaction) error {
        dateDocRef := f.documentRef(hourTime)

        firebaseDate, err := f.getDateDTO(
            // getDateDTO should be used both for transactional and non transactional query,
            // the best way for that is to use closure
            func() (doc *firestore.DocumentSnapshot, err error) {
                return transaction.Get(dateDocRef)
            },
            hourTime,
        )
        if err != nil {
            return err
        }

        hourFromDB, err := f.domainHourFromDateDTO(firebaseDate, hourTime)
        if err != nil {
            return err
        }

        updatedHour, err := updateFn(hourFromDB)
        if err != nil {
            return errors.Wrap(err, "unable to update hour")
        }
        updateHourInDataDTO(updatedHour, &firebaseDate)

        return transaction.Set(dateDocRef, firebaseDate)
    })

    return errors.Wrap(err, "firestore transaction failed")
}
```

### 3g. SQL schema and upsert pattern

```sql
CREATE TABLE `hours`
(
    hour         TIMESTAMP                                                 NOT NULL,
    availability ENUM ('available', 'not_available', 'training_scheduled') NOT NULL,
    PRIMARY KEY (hour)
);
```

```go
func (m MySQLHourRepository) upsertHour(tx *sqlx.Tx, hourToUpdate *hour.Hour) error {
    updatedDbHour := mysqlHour{
        Hour:         hourToUpdate.Time().UTC(),
        Availability: hourToUpdate.Availability().String(),
    }

    _, err := tx.NamedExec(
        `INSERT INTO
            hours (hour, availability)
        VALUES
            (:hour, :availability)
        ON DUPLICATE KEY UPDATE
            availability = :availability`,
        updatedDbHour,
    )
    if err != nil {
        return errors.Wrap(err, "unable to upsert hour")
    }
    return nil
}
```

### 3h. Swappability is the test

If you can replace Firestore with MySQL by implementing the same interface, the repository is correctly abstracted. The book demonstrates exactly this: both Firestore and MySQL implementations exist for `hour.Repository`, and the same test suite runs against both. Can you guess what database the handler uses? No — and that is the point.

---

## 4. Clean Architecture (Ports, Adapters, Application, Domain)

### 4a. Four layers, four packages

```
internal/trainings/
├── domain/         # pure business logic, no dependencies on other layers
├── app/            # use cases (commands + queries), depends on domain
│   ├── command/
│   └── query/
├── adapters/       # DB, gRPC clients, Pub/Sub — depends on app + domain
└── ports/          # HTTP handlers, gRPC servers — depends on app only
```

Note on naming: the book's `ports` equals "Primary Adapters" in Hexagonal Architecture. The book's `adapters` equals "Secondary Adapters." The book chose different names because primary/secondary is hard to grasp. Go's implicit interfaces mean no dedicated "ports" interface package is needed.

If the project grows, add subdirectories: `adapters/hour/mysql_repository.go` or `ports/http/hour_handler.go`.

### 4b. Dependency direction — the one rule

```
ports    ──> app ──> domain
adapters ──> app ──> domain
```

- **Domain** knows nothing about other layers whatsoever — contains pure business logic.
- **App** can import domain; knows nothing about outer layers. It cannot tell if called from HTTP, gRPC, or CLI.
- **Ports** can import app; are the entry points, so they execute commands/queries. They cannot directly access adapters.
- **Adapters** can import app and domain; operate on domain types, retrieve them from the database.

This is the Dependency Inversion Principle (the "D" in SOLID) applied at the package level. Go's implicit interfaces make it natural — no "implements" declarations needed. Import cycles enforce the rule mechanically.

### 4c. Define interfaces where they are used, not in a separate package

The application layer defines what it needs; adapters implement it. Because Go interfaces don't need to be explicitly implemented, define them next to the code that needs them:

```go
// app/command/training_service.go — interfaces defined in the app layer
type trainingRepository interface {
    AddTraining(ctx context.Context, tr *training.Training) error
    GetTraining(ctx context.Context, trainingUUID string, user training.User) (*training.Training, error)
    UpdateTraining(
        ctx context.Context,
        trainingUUID string,
        user training.User,
        updateFn func(ctx context.Context, tr *training.Training) (*training.Training, error),
    ) error
}

type userService interface {
    UpdateTrainingBalance(ctx context.Context, userID string, amountChange int) error
}

type trainerService interface {
    ScheduleTraining(ctx context.Context, trainingTime time.Time) error
    CancelTraining(ctx context.Context, trainingTime time.Time) error
    MoveTraining(ctx context.Context, newTime time.Time, originalTime time.Time) error
}
```

Note that "user" and "trainer" are application (business) concepts — not microservices. It just happens they live in microservices with the same names.

### 4d. Dependency injection in `main.go`

```go
trainingsRepository := adapters.NewTrainingsFirestoreRepository(client)
trainerGrpc := adapters.NewTrainerGrpc(trainerClient)
usersGrpc := adapters.NewUsersGrpc(usersClient)

trainingsService := app.NewTrainingsService(trainingsRepository, trainerGrpc, usersGrpc)
```

Panic in constructors on nil dependencies — fail fast is better than nil pointer panics in production:

```go
func NewTrainingsService(
    repo trainingRepository,
    trainerService trainerService,
    userService userService,
) TrainingService {
    if repo == nil {
        panic("missing trainingRepository")
    }
    if trainerService == nil {
        panic("missing trainerService")
    }
    if userService == nil {
        panic("missing userService")
    }
    return TrainingService{
        repo:           repo,
        trainerService: trainerService,
        userService:    userService,
    }
}
```

### 4e. Application layer is thin orchestration

If you can't tell from reading application layer code what database it uses or what HTTP endpoint it calls, that's good design. All `if`s expressing business rules belong in the domain layer. When you start seeing `if`s related to logic in your application layer, move them to the domain layer.

The application layer's job: get entity from repository, orchestrate domain stuff, call external services, return entity to be saved.

### 4f. Port-agnostic errors with slugs

Return typed errors with slugs from the application layer so both HTTP and gRPC ports can handle them appropriately without the app knowing which protocol is in use:

```go
return nil, errors.NewIncorrectInputError("date-from-after-date-to", "Date from after date to")
// HTTP port: translates to 400 Bad Request
// gRPC port: translates to codes.InvalidArgument
```

The slug (`"date-from-after-date-to"`) can be translated on the frontend and shown to the user. This prevents leaking implementation details to the application logic.

### 4g. Apply Clean Architecture where it makes sense

The `users` microservice in Wild Workouts is tiny with no application logic. The book explicitly skips Clean Architecture there: "As with every technique, apply Clean Architecture where it makes sense." Don't force the pattern on a simple CRUD service.

### 4h. Enforce with static analysis

Use [go-cleanarch](https://github.com/roblaszczak/go-cleanarch) in CI to prevent adapters from importing ports, or domain from importing anything outside itself. This catches architectural violations automatically without relying on code review.

---

## 5. CQRS

### 5a. Problems CQRS solves

CQRS solves three problems the book explicitly names:
1. One big, unmaintainable model that is hard to understand and change
2. Limited parallelism in development — multiple people can't work on the same service simultaneously
3. Scaling issues — read and write paths can't be optimised independently

Without proper planning, splitting to more microservices often makes these problems *worse*: complex logic becomes harder to understand when distributed, distributed transactions are far more complex than single-DB transactions, and coordinating changes across team-owned services is slower.

### 5b. Commands vs Queries — the fundamental split

- **Command**: changes system state, returns only an error (no data). Named return `(err error)` for deferred logging.
- **Query**: reads data, changes nothing, returns data. Side effects like logs and metrics do not count as "modifying."

Keep them as separate structs with a `Handle` method. The book: "CQRS is a very simple pattern that doesn't require a lot of investment. It can be easily extended with more complex techniques like event-driven architecture, event-sourcing, or polyglot persistence. But they're not always needed."

### 5c. `ApproveTrainingRescheduleHandler` — full example

The pre-CQRS implementation in `TrainingService.ApproveTrainingReschedule` had magic validations directly in the application service. It also missed calling the external trainer service to move the training. After CQRS + DDD refactoring:

```go
// app/command/approve_training_reschedule.go
package command

type ApproveTrainingReschedule struct {
    TrainingUUID string
    User         training.User // use domain types, not raw strings — type safety at compile time
}

type ApproveTrainingRescheduleHandler struct {
    repo           training.Repository
    userService    UserService
    trainerService TrainerService
}

func (h ApproveTrainingRescheduleHandler) Handle(ctx context.Context, cmd ApproveTrainingReschedule) (err error) {
    defer func() {
        logs.LogCommandExecution("ApproveTrainingReschedule", cmd, err)
    }()

    return h.repo.UpdateTraining(
        ctx,
        cmd.TrainingUUID,
        cmd.User,
        func(ctx context.Context, tr *training.Training) (*training.Training, error) {
            originalTrainingTime := tr.Time()

            if err := tr.ApproveReschedule(cmd.User.Type()); err != nil {
                return nil, err
            }

            err := h.trainerService.MoveTraining(ctx, tr.Time(), originalTrainingTime)
            if err != nil {
                return nil, err
            }

            return tr, nil
        },
    )
}
```

No domain logic in the handler. All validation lives in `tr.ApproveReschedule`. The `defer` with named return `(err error)` ensures logging fires on both success and failure.

### 5d. `CancelTrainingHandler` — domain orchestration with `CancelBalanceDelta`

```go
func (h CancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) (err error) {
    defer func() {
        logs.LogCommandExecution("CancelTrainingHandler", cmd, err)
    }()

    return h.repo.UpdateTraining(
        ctx,
        cmd.TrainingUUID,
        cmd.User,
        func(ctx context.Context, tr *training.Training) (*training.Training, error) {
            if err := tr.Cancel(); err != nil {
                return nil, err
            }

            if balanceDelta := training.CancelBalanceDelta(*tr, cmd.User.Type()); balanceDelta != 0 {
                err := h.userService.UpdateTrainingBalance(ctx, tr.UserUUID(), balanceDelta)
                if err != nil {
                    return nil, errors.Wrap(err, "unable to change trainings balance")
                }
            }

            if err := h.trainerService.CancelTraining(ctx, tr.Time()); err != nil {
                return nil, errors.Wrap(err, "unable to cancel training")
            }

            return tr, nil
        },
    )
}
```

Flow: cancel the domain entity, adjust balance via user service, notify trainer service. No domain logic in the handler — `Cancel()` and `CancelBalanceDelta()` are in the domain layer.

### 5e. Query structure — `AvailableHoursHandler`

```go
package query

type AvailableHoursReadModel interface {
    AvailableHours(ctx context.Context, from time.Time, to time.Time) ([]Date, error)
}

type AvailableHoursHandler struct {
    readModel AvailableHoursReadModel
}

type AvailableHours struct {
    From time.Time
    To   time.Time
}

func (h AvailableHoursHandler) Handle(ctx context.Context, query AvailableHours) (d []Date, err error) {
    start := time.Now()
    defer func() {
        logrus.
            WithError(err).
            WithField("duration", time.Since(start)).
            Debug("AvailableHoursHandler executed")
    }()

    if query.From.After(query.To) {
        return nil, errors.NewIncorrectInputError("date-from-after-date-to", "Date from after date to")
    }

    return h.readModel.AvailableHours(ctx, query.From, query.To)
}
```

Commands and queries are "a great place for all cross-cutting concerns, like logging and instrumentation. Thanks to putting that here, we are sure that performance is measured in the same way whether it's called from HTTP or gRPC port."

### 5f. Query model types — not domain types, not OpenAPI types

Define separate query model types shaped by UI requirements. Don't use domain types or OpenAPI-generated types directly in queries:

```go
package query

type Date struct {
    Date         time.Time
    HasFreeHours bool
    Hours        []Hour
}

type Hour struct {
    Available            bool
    HasTrainingScheduled bool
    Hour                 time.Time
}
```

The `query.Date` type is more complex than `hour.Hour` because it's driven by UI requirements (`HasFreeHours` for display). As the application grows, differences between domain and query models grow. Separation allows independent evolution.

For simple queries, you can use the domain repository directly instead of a dedicated read model interface:

```go
type HourAvailabilityHandler struct {
    hourRepo hour.Repository
}

func (h HourAvailabilityHandler) Handle(ctx context.Context, time time.Time) (bool, error) {
    hour, err := h.hourRepo.GetHour(ctx, time)
    if err != nil {
        return false, err
    }
    return hour.IsAvailable(), nil
}
```

### 5g. Collect commands and queries in the `Application` struct

```go
package app

type Application struct {
    Commands Commands
    Queries  Queries
}

type Commands struct {
    ApproveTrainingReschedule command.ApproveTrainingRescheduleHandler
    CancelTraining            command.CancelTrainingHandler
    ScheduleTraining          command.ScheduleTrainingHandler
    RequestTrainingReschedule command.RequestTrainingRescheduleHandler
    RejectTrainingReschedule  command.RejectTrainingRescheduleHandler
}

type Queries struct {
    AvailableHours   query.AvailableHoursHandler
    HourAvailability query.HourAvailabilityHandler
    AllTrainings     query.AllTrainingsHandler
    TrainingsForUser query.TrainingsForUserHandler
}
```

HTTP and gRPC ports inject this struct and call handlers from it. The command handler can be called from any port — HTTP, gRPC, CLI, Pub/Sub, or migration scripts:

```go
type HttpServer struct {
    app app.Application
}

func (h HttpServer) ApproveRescheduleTraining(w http.ResponseWriter, r *http.Request) {
    trainingUUID := chi.URLParam(r, "trainingUUID")
    user, err := newDomainUserFromAuthUser(r.Context())
    if err != nil {
        httperr.RespondWithSlugError(err, w, r)
        return
    }

    err = h.app.Commands.ApproveTrainingReschedule.Handle(r.Context(), command.ApproveTrainingReschedule{
        User:         user,
        TrainingUUID: trainingUUID,
    })
    if err != nil {
        httperr.RespondWithSlugError(err, w, r)
        return
    }
}
```

### 5h. Naming: use ubiquitous language

Avoid `CreateTraining`, `DeleteTraining`. Use `ScheduleTraining`, `CancelTraining`. The test: "Think twice if any of your command names really need to start with Create/Delete/Update." Go to your business stakeholders and listen to how they call operations.

If CQRS is the standard in your team, a new engineer can look at the list of commands and queries and immediately understand what the service does, without jumping through random places in code.

### 5i. Returning the created resource — UUID in the port

For POST endpoints that need to return the new resource: generate the UUID in the port and pass it in the command. Return `204 No Content` with `Content-Location` header. The client fetches if needed. This approach also works when the command handler is asynchronous:

```go
cmd := command.ScheduleTraining{
    TrainingUUID: uuid.New().String(), // generated in the port, passed to the command
    UserUUID:     user.UUID,
    UserName:     user.DisplayName,
    TrainingTime: postTraining.Time,
    Notes:        postTraining.Notes,
}
err = h.app.Commands.ScheduleTraining.Handle(r.Context(), cmd)
if err != nil {
    httperr.RespondWithSlugError(err, w, r)
    return
}

w.Header().Set("content-location", "/trainings/"+cmd.TrainingUUID)
w.WriteHeader(http.StatusNoContent)
```

As a last resort you can return the ID from the handler — "it won't be the end of the world."

### 5j. When NOT to use CQRS

Simple CRUD services where reads and writes use identical models. The `users` microservice in Wild Workouts just stores and retrieves profiles — no application logic. Forcing CQRS there would be unnecessary overhead.

But watch simple services: "if you notice the logic grows and development is painful, maybe it's time for some refactoring?"

### 5k. Simple command handlers — don't skip the layer

Some command handlers are very simple. Don't skip the application layer to save boilerplate. Writing this type takes 3 minutes; people who maintain and extend this code later will appreciate that effort:

```go
func (h RequestTrainingRescheduleHandler) Handle(ctx context.Context, cmd RequestTrainingReschedule) (err error) {
    defer func() {
        logs.LogCommandExecution("RequestTrainingReschedule", cmd, err)
    }()

    return h.repo.UpdateTraining(
        ctx,
        cmd.TrainingUUID,
        cmd.User,
        func(ctx context.Context, tr *training.Training) (*training.Training, error) {
            if err := tr.UpdateNotes(cmd.NewNotes); err != nil {
                return nil, err
            }

            tr.ProposeReschedule(cmd.NewTime, cmd.User.Type())

            return tr, nil
        },
    )
}
```

### 5l. Extending CQRS — future patterns

CQRS enables extension with:

- **Async commands**: Use a message queue (e.g., Watermill) to execute commands asynchronously for slow operations (external calls, heavy computation). This has additional infrastructure requirements but the command interface stays the same.

- **Polyglot persistence**: The current read and write paths use the same database. For more complex queries or high read throughput, duplicate data into a read-optimised store (e.g., Elasticsearch). Synchronise via events. Key tradeoff: eventual consistency. Defer this until you have evidence of the need.

- **Event Sourcing**: Instead of storing current model state, store a list of events used by writes. CQRS naturally separates the read model so queries work independently. Provides out-of-the-box audit trail. Recommended when working in domains with strict audit requirements (e.g., financial systems). Reference: Greg Young's *Versioning in an Event Sourced System*.

Start without these extensions. They have infrastructure and operational complexity costs. Defer these decisions.

---

## 6. Repository Secure by Design

### 6a. The motivating bug

The book cites a real security bug in the Harbor project: a regular user could create an admin user because the authorization check was missing from one handler. The fix was one `if` statement. But did it secure the application from a *similar* bug in the future? No. The same mistake could be made again by anyone unfamiliar with the codebase.

Good design should not *allow* using code in an invalid way. The goal: authorization that is structurally impossible to bypass.

### 6b. `CanUserSeeTraining` in the domain layer

Put the authorization *logic* in the domain layer (testable, reusable, readable by business). Put the *enforcement* in the repository (unavoidable by any caller):

```go
// domain/training/user.go
type ForbiddenToSeeTrainingError struct {
    RequestingUserUUID string
    TrainingOwnerUUID  string
}

func (f ForbiddenToSeeTrainingError) Error() string {
    return fmt.Sprintf(
        "user '%s' can't see user '%s' training",
        f.RequestingUserUUID, f.TrainingOwnerUUID,
    )
}

func CanUserSeeTraining(user User, training Training) error {
    if user.Type() == Trainer {
        return nil
    }
    if user.UUID() == training.UserUUID() {
        return nil
    }
    return ForbiddenToSeeTrainingError{user.UUID(), training.UserUUID()}
}
```

### 6c. Enforce in `GetTraining` and `UpdateTraining`

```go
func (r TrainingsFirestoreRepository) GetTraining(
    ctx context.Context,
    trainingUUID string,
    user training.User, // explicit parameter — cannot be bypassed
) (*training.Training, error) {
    firestoreTraining, err := r.trainingsCollection().Doc(trainingUUID).Get(ctx)
    if status.Code(err) == codes.NotFound {
        return nil, training.NotFoundError{trainingUUID}
    }
    if err != nil {
        return nil, errors.Wrap(err, "unable to get actual docs")
    }

    tr, err := r.unmarshalTraining(firestoreTraining)
    if err != nil {
        return nil, err
    }

    if err := training.CanUserSeeTraining(user, *tr); err != nil {
        return nil, err
    }

    return tr, nil
}
```

Same pattern in `UpdateTraining` — `CanUserSeeTraining` is called before `updateFn` executes:

```go
func (r TrainingsFirestoreRepository) UpdateTraining(
    ctx context.Context,
    trainingUUID string,
    user training.User,
    updateFn func(ctx context.Context, tr *training.Training) (*training.Training, error),
) error {
    return r.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        documentRef := trainingsCollection.Doc(trainingUUID)
        firestoreTraining, err := tx.Get(documentRef)
        // ... fetch and unmarshal tr ...

        if err := training.CanUserSeeTraining(user, *tr); err != nil {
            return err
        }

        updatedTraining, err := updateFn(ctx, tr)
        if err != nil {
            return err
        }

        return tx.Set(documentRef, r.marshalTraining(updatedTraining))
    })
}
```

### 6d. Never pass auth via `context.Context`

Passing authentication details through `context.Context` loses static typing and hides the function's true inputs:

```go
// BAD: auth hidden in context — breaks static typing, hides inputs
func (r Repo) GetTraining(ctx context.Context, trainingUUID string) (*training.Training, error)

// GOOD: user is explicit — type-safe, self-documenting
func (r Repo) GetTraining(ctx context.Context, trainingUUID string, user training.User) (*training.Training, error)
```

`context.Context` is for cancellation, deadlines, and tracing — not for carrying business data or authentication. Passing values via context may be a symptom of bad design: "Maybe the function is doing too much, and it's hard to pass all arguments there? Perhaps it's the time to decompose that?"

### 6e. `UpdateTrainingByMigration` for internal/admin operations

For background jobs, migrations, or admin operations that skip user-scoped access control, create explicitly named methods:

```go
// Clearly signals no user context — caller knows the security implications
func (r TrainingsFirestoreRepository) UpdateTrainingByMigration(
    ctx context.Context,
    trainingUUID string,
    updateFn func(ctx context.Context, tr *training.Training) (*training.Training, error),
) error
```

Never use a "fake user" or role hacks to bypass access checks. From the book's experience, fake users lead to a lot of extra `if` statements polluting the code and obfuscate the audit log.

### 6f. Collections: filter in the database, not in a loop

For collection access, filter by query parameters rather than calling `CanUserSeeTraining` in a loop — that is expensive and slow:

```go
func (r TrainingsFirestoreRepository) FindTrainingsForUser(ctx context.Context, userUUID string) ([]query.Training, error) {
    query := r.trainingsCollection().Query.
        Where("Time", ">=", time.Now().Add(-time.Hour*24)).
        Where("UserUuid", "==", userUUID).
        Where("Canceled", "==", false)

    iter := query.Documents(ctx)
    return r.trainingModelsToQuery(iter)
}
```

---

## 7. Tests Architecture

### 7a. Four principles of high-quality tests

Before writing tests, ensure your test suite meets these four criteria:

1. **Fast** — tests must be fast; no compromise. Context switching caused by long CI waits is a productivity killer. Target under 1 minute locally; ideally under 10 seconds. The Pareto principle applies: find the 20% of slowest tests causing 80% of the wait.

2. **Testing enough scenarios on all levels** — 70–80% coverage is a good result in Go. Don't chase 100%. Ask "how easily can this break?" Tests that mirror the implementation rather than specifying behaviour have low value. Use multiple levels (unit, integration, component, E2E); tests on adjacent levels should overlap to verify integration is correct. Note: applications that aggregate data may need a reversed test pyramid (more integration tests than unit tests) because most code is database-related.

3. **Robust and deterministic** — flaky tests are worse than no tests. They erode trust and discourage adding new tests. Fix flakes immediately — broken windows theory applies.

4. **Executable locally** — developers must be able to run the full suite locally before pushing. E2E tests are the exception: they verify contracts, not logic. Use gRPC, protobuf, or OpenAPI to make contracts robust and independently testable.

### 7b. Four kinds of tests — the team reference table

Create a shared table for your team so definitions are unambiguous and unproductive discussions about test categories are eliminated:

| Feature | Unit | Integration | Component | End-to-End |
|---|---|---|---|---|
| Docker database | No | Yes | Yes | Yes |
| External services | No | No | No | Yes |
| Uses mocks | Most deps | Usually none | External services only | None |
| Tests API | Go package | Go package | HTTP and gRPC | HTTP |
| Focused on business cases | Depends | No | Yes | Yes |

### 7c. Unit tests: domain layer

The domain layer has no external dependencies — tests are fast and need no mocks. Aim for high coverage here. Use table-driven tests for all corner cases. Use the `_test` package suffix to enforce black-box testing (only public API).

```go
func TestHour_ScheduleTraining(t *testing.T) {
    h, err := hour.NewAvailableHour(validTrainingHour())
    require.NoError(t, err)

    require.NoError(t, h.ScheduleTraining())

    assert.True(t, h.HasTrainingScheduled())
    assert.False(t, h.IsAvailable())
}

func TestHour_ScheduleTraining_with_not_available(t *testing.T) {
    h := newNotAvailableHour(t)
    assert.Equal(t, hour.ErrHourNotAvailable, h.ScheduleTraining())
}
```

Table-driven tests for domain validation:

```go
func TestFactoryConfig_Validate(t *testing.T) {
    testCases := []struct {
        Name        string
        Config      hour.FactoryConfig
        ExpectedErr string
    }{
        {
            Name: "valid",
            Config: hour.FactoryConfig{
                MaxWeeksInTheFutureToSet: 10,
                MinUtcHour:               10,
                MaxUtcHour:               12,
            },
            ExpectedErr: "",
        },
        // ... more cases
    }

    for _, c := range testCases {
        t.Run(c.Name, func(t *testing.T) {
            err := c.Config.Validate()
            if c.ExpectedErr != "" {
                assert.EqualError(t, err, c.ExpectedErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 7d. Unit tests: application layer

Mock dependencies with simple hand-written structs — no mocking library needed. Only test complex orchestration. Skip application tests where code just delegates to the domain with no logic. Watch out for tests that only test mocks:

```go
type trainerServiceMock struct {
    trainingsCancelled []time.Time
}

func (t *trainerServiceMock) CancelTraining(ctx context.Context, trainingTime time.Time) error {
    t.trainingsCancelled = append(t.trainingsCancelled, trainingTime)
    return nil
}

func newDependencies() dependencies {
    repository := &repositoryMock{}
    trainerService := &trainerServiceMock{}
    userService := &userServiceMock{}

    return dependencies{
        repository:     repository,
        trainerService: trainerService,
        userService:    userService,
        handler:        command.NewCancelTrainingHandler(repository, userService, trainerService),
    }
}
```

Use table-driven tests to make all business rule cases easy to read. The `CancelTraining` handler test shows the balance-change rules:

```go
{
    Name:     "return_training_balance_when_trainer_cancels",
    UserRole: "trainer",
    Training: app.Training{
        UserUUID: "trainer-id",
        Time:     time.Now().Add(48 * time.Hour),
    },
    ShouldUpdateBalance:   true,
    ExpectedBalanceChange: 1,
},
{
    Name:     "extra_training_balance_when_trainer_cancels_before_24h",
    UserRole: "trainer",
    Training: app.Training{
        UserUUID: "trainer-id",
        Time:     time.Now().Add(12 * time.Hour),
    },
    ShouldUpdateBalance:   true,
    ExpectedBalanceChange: 2,
},
```

Before introducing Clean Architecture and CQRS, these tests required a real database and all services running. After the refactor, they are pure unit tests — fast and fully local.

### 7e. `require` vs `assert` — stop vs collect

Use `require` when a failing assertion makes the rest of the test meaningless (nil would cause a panic). Use `assert` to collect all failures in one run:

```go
h, err := hour.NewAvailableHour(validTrainingHour())
require.NoError(t, err)           // stop — if Hour is nil, next lines panic

require.NoError(t, h.ScheduleTraining())

assert.True(t, h.HasTrainingScheduled())  // continue — collect both failures
assert.False(t, h.IsAvailable())
```

`require` calls `t.FailNow`. `assert` marks failure and continues. Use `require` at the top for setup and for errors where nil causes a panic. Use `assert` for the actual assertions.

### 7f. Integration tests: one test suite for all repository implementations

Run against real infrastructure (Docker). Test all repository implementations with the same test suite by parameterising over them. Run in parallel — target less than **200ms** total. All tests use `_test` package suffix (black-box testing):

```go
package main_test

func TestRepository(t *testing.T) {
    rand.Seed(time.Now().UTC().UnixNano())

    repositories := createRepositories(t)

    for i := range repositories {
        // capture loop variable — required in Go 1.21 and earlier
        r := repositories[i]

        t.Run(r.Name, func(t *testing.T) {
            t.Parallel()

            t.Run("testUpdateHour", func(t *testing.T) {
                t.Parallel()
                testUpdateHour(t, r.Repository)
            })
            t.Run("testUpdateHour_parallel", func(t *testing.T) {
                t.Parallel()
                testUpdateHour_parallel(t, r.Repository)
            })
            t.Run("testHourRepository_update_existing", func(t *testing.T) {
                t.Parallel()
                testHourRepository_update_existing(t, r.Repository)
            })
            t.Run("testUpdateHour_rollback", func(t *testing.T) {
                t.Parallel()
                testUpdateHour_rollback(t, r.Repository)
            })
        })
    }
}
```

The `testUpdateHour` function uses a test table — each case creates a different starting state and saves it through `UpdateHour`:

```go
func testUpdateHour(t *testing.T, repository hour.Repository) {
    t.Helper()
    ctx := context.Background()

    testCases := []struct {
        Name       string
        CreateHour func(*testing.T) *hour.Hour
    }{
        {
            Name: "available_hour",
            CreateHour: func(t *testing.T) *hour.Hour {
                return newValidAvailableHour(t)
            },
        },
        {
            Name: "not_available_hour",
            CreateHour: func(t *testing.T) *hour.Hour {
                h := newValidAvailableHour(t)
                require.NoError(t, h.MakeNotAvailable())
                return h
            },
        },
        {
            Name: "hour_with_training",
            CreateHour: func(t *testing.T) *hour.Hour {
                h := newValidAvailableHour(t)
                require.NoError(t, h.ScheduleTraining())
                return h
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.Name, func(t *testing.T) {
            newHour := tc.CreateHour(t)

            err := repository.UpdateHour(ctx, newHour.Time(), func(_ *hour.Hour) (*hour.Hour, error) {
                return newHour, nil
            })
            require.NoError(t, err)

            assertHourInRepository(ctx, t, repository, newHour)
        })
    }
}
```

The `assertHourInRepository` helper reads back the saved hour and compares it, removing assertion duplication:

```go
func assertHourInRepository(ctx context.Context, t *testing.T, repo hour.Repository, hour *hour.Hour) {
    require.NotNil(t, hour)

    hourFromRepo, err := repo.GetOrCreateHour(ctx, hour.Time())
    require.NoError(t, err)

    assert.Equal(t, hour, hourFromRepo)
}
```

### 7g. Test the rollback explicitly — `testUpdateHour_rollback`

Verify that when `updateFn` returns an error, the database is left unchanged. This test validates the transaction is actually rolling back:

```go
func testUpdateHour_rollback(t *testing.T, repository hour.Repository) {
    t.Helper()
    ctx := context.Background()

    hourTime := newValidHourTime()

    // First: make it available
    err := repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
        require.NoError(t, h.MakeAvailable())
        return h, nil
    })
    require.NoError(t, err)

    // Second: attempt to make it not-available, but return error from updateFn
    err = repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
        assert.True(t, h.IsAvailable())
        require.NoError(t, h.MakeNotAvailable())
        return h, errors.New("something went wrong")
    })
    require.Error(t, err)

    // The change should have been rolled back
    persistedHour, err := repository.GetOrCreateHour(ctx, hourTime)
    require.NoError(t, err)
    assert.True(t, persistedHour.IsAvailable(), "availability change was persisted, not rolled back")
}
```

### 7h. Tests sabotage technique

After writing a test, deliberately break the implementation to verify the test actually catches the bug. If the test still passes with a broken implementation, it is not testing what you think:

```go
// Temporarily break finishTransaction — always commit regardless of error:
func (m MySQLHourRepository) finishTransaction(err error, tx *sqlx.Tx) error {
    if commitErr := tx.Commit(); commitErr != nil {
        return errors.Wrap(err, "failed to commit tx")
    }
    return nil
}
// Now run tests — if testUpdateHour_rollback still passes, the test is wrong.
```

This is especially important for integration tests where false positives are costly.

### 7i. Race condition test — `testUpdateHour_parallel`

20 goroutines released simultaneously to attempt the same update. Only one should succeed. `close(startWorkers)` unblocks all goroutines at once to maximise the chance of a real race condition:

```go
func testUpdateHour_parallel(t *testing.T, repository hour.Repository) {
    workersCount := 20
    workersDone := sync.WaitGroup{}
    workersDone.Add(workersCount)

    // closing startWorkers will unblock all workers at once,
    // thanks to that it will be more likely to have race condition
    startWorkers := make(chan struct{})
    trainingsScheduled := make(chan int, workersCount)

    for worker := 0; worker < workersCount; worker++ {
        workerNum := worker
        go func() {
            defer workersDone.Done()
            <-startWorkers // wait for simultaneous release

            schedulingTraining := false

            err := repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
                if h.HasTrainingScheduled() {
                    return h, nil
                }
                if err := h.ScheduleTraining(); err != nil {
                    return nil, err
                }
                schedulingTraining = true
                return h, nil
            })

            if schedulingTraining && err == nil {
                trainingsScheduled <- workerNum
            }
        }()
    }

    close(startWorkers) // release all goroutines at once
    workersDone.Wait()
    close(trainingsScheduled)

    var workersScheduledTraining []int
    for workerNum := range trainingsScheduled {
        workersScheduledTraining = append(workersScheduledTraining, workerNum)
    }

    assert.Len(t, workersScheduledTraining, 1, "only one worker should schedule training")
}
```

### 7j. Test isolation: `newValidHourTime()` with `sync.Map`

Never share test data between tests. Generate a unique time/ID per test run using a `sync.Map` — this avoids the need for cleanup and allows full parallelism:

```go
// usedHours stores hours used during the test run,
// to ensure that within one test run we are not using the same hour
var usedHours = sync.Map{}

func newValidHourTime() time.Time {
    for {
        minTime := time.Now().AddDate(0, 0, 1)
        minTimestamp := minTime.Unix()
        maxTimestamp := minTime.AddDate(0, 0, testHourFactory.Config().MaxWeeksInTheFutureToSet*7).Unix()

        t := time.Unix(rand.Int63n(maxTimestamp-minTimestamp)+minTimestamp, 0).Truncate(time.Hour).Local()

        _, alreadyUsed := usedHours.LoadOrStore(t.Unix(), true)
        if !alreadyUsed {
            return t
        }
    }
}
```

**Why no cleanup:** cleaning up requires knowing what was created. Downsides of cleanup: when it doesn't work correctly it creates hard-to-debug issues; it makes tests slower; it adds maintenance overhead; it makes parallel tests harder. Prefer unique IDs.

Note: some scenarios cannot be parallelised (pagination, global counters). Keep those tests as short as possible.

### 7k. Never use sleep in tests

Never use `time.Sleep` to wait for async side effects. Use `assert.Eventually` instead:

```go
assert.Eventually(
    t,
    func() bool { return true }, // condition
    time.Second,                  // waitFor
    10*time.Millisecond,          // tick
)
```

For synchronisation, use channels or `sync.WaitGroup{}` — they are faster and more stable than sleep.

### 7l. Component tests: single service in isolation

Spin up the real HTTP/gRPC server in a goroutine. Use real infrastructure (Docker DB). Mock only external service adapters (gRPC clients to other microservices). Test the full request path from port to database and back.

Two constructors for the application: one for production, one for component tests:

```go
func NewApplication(ctx context.Context) (app.Application, func()) {
    trainerGrpc := adapters.NewTrainerGrpc(trainerClient)
    usersGrpc := adapters.NewUsersGrpc(usersClient)
    return newApplication(ctx, trainerGrpc, usersGrpc), cleanupFn
}

func NewComponentTestApplication(ctx context.Context) app.Application {
    return newApplication(ctx, TrainerServiceMock{}, UserServiceMock{})
}
```

Use `TestMain` and `tests.WaitForPort` — never sleep waiting for the server to start. Start one instance for all component tests — rebuilding per test is too slow:

```go
func startService() bool {
    app := NewComponentTestApplication(context.Background())

    trainingsHTTPAddr := os.Getenv("TRAININGS_HTTP_ADDR")
    go server.RunHTTPServerOnAddr(trainingsHTTPAddr, func(router chi.Router) http.Handler {
        return ports.HandlerFromMux(ports.NewHttpServer(app), router)
    })

    ok := tests.WaitForPort(trainingsHTTPAddr)
    if !ok {
        log.Println("Timed out waiting for trainings HTTP to come up")
    }
    return ok
}

func TestMain(m *testing.M) {
    if !startService() {
        os.Exit(1)
    }
    os.Exit(m.Run())
}
```

Use generated OpenAPI client wrappers, not raw HTTP calls. Wrap them in test helper methods:

```go
func (c TrainingsHTTPClient) CreateTraining(t *testing.T, note string, hour time.Time) string {
    response, err := c.client.CreateTrainingWithResponse(context.Background(), trainings.CreateTrainingJSONRequestBody{
        Notes: note,
        Time:  hour,
    })
    require.NoError(t, err)
    require.Equal(t, http.StatusNoContent, response.StatusCode())

    contentLocation := response.HTTPResponse.Header.Get("content-location")
    return lastPathElement(contentLocation)
}

// caller — single readable line:
trainingUUID := client.CreateTraining(t, "some note", hour)
```

What to test in component tests: the happy path and basic error cases. Don't check corner cases — unit and integration tests already cover those.

### 7m. E2E tests: test contracts, not logic

E2E tests spin up all services via docker-compose and verify a few critical paths. They verify contracts (does service A respond correctly to service B's request?), not business logic (that is covered by unit and component tests):

```go
// E2E: happy path only
user := usersHTTPClient.GetCurrentUser(t)
originalBalance := user.Balance

trainingUUID := trainingsHTTPClient.CreateTraining(t, "some note", hour)

trainingsResponse := trainingsHTTPClient.GetTrainings(t)
require.Len(t, trainingsResponse.Trainings, 1)
require.Equal(t, trainingUUID, trainingsResponse.Trainings[0].Uuid)

user = usersHTTPClient.GetCurrentUser(t)
require.Equal(t, originalBalance, user.Balance, "balance should be updated after scheduling")
```

Keep the E2E suite small — only cross-service integration paths. "These tests work as a double-check. They shouldn't fail most of the time, and if they do, it usually means someone has broken the contract."

The Accelerate research finding on best teams: "We can do most of our testing without requiring an integrated environment. We can and do deploy or release our application independently of other applications/services it depends on."

### 7n. Custom assertion helpers and `go-cmp`

Extract repeated assertion patterns into named helpers:

```go
func assertTrainingsEquals(t *testing.T, tr1, tr2 *training.Training) {
    cmpOpts := []cmp.Option{
        cmpRoundTimeOpt,
        cmp.AllowUnexported(
            training.UserType{},
            time.Time{},
            training.Training{},
        ),
    }

    assert.True(
        t,
        cmp.Equal(tr1, tr2, cmpOpts...),
        cmp.Diff(tr1, tr2, cmpOpts...),
    )
}
```

Use `github.com/google/go-cmp` with `cmp.AllowUnexported` instead of `reflect.DeepEqual` for domain types with private fields. `reflect.DeepEqual` gives wrong results on unexported fields in some cases; `go-cmp` handles them correctly and produces readable diffs.

### 7o. Loop variable capture (Go 1.21 and earlier)

When calling `t.Parallel()` inside `range` loops, capture the loop variable with the index approach:

```go
// BAD in Go 1.21 and earlier — all subtests run with the last value
for _, c := range testCases {
    t.Run(c.Name, func(t *testing.T) {
        t.Parallel()
        // c is the last element!
    })
}

// GOOD — explicit index access (preferred over c := c)
for i := range testCases {
    c := testCases[i]
    t.Run(c.Name, func(t *testing.T) {
        t.Parallel()
        // use c here
    })
}
```

Go 1.22+ fixes this automatically. The book preferred the index approach over `c := c` as it's more obvious to readers.

### 7p. Test coverage target and database libraries

**Coverage target:** 70–80% is a healthy result in Go. Don't chase 100%. Ask "how easily can this break?" Tests that replicate implementation rather than specify behaviour have low value.

**`sqlx` vs SQLBoiler:**
- `sqlx`: thin wrapper over `database/sql`. Use for simpler data models or when you want full SQL control.
- `SQLBoiler`: generates full type-safe models from the database schema. Use when you have many tables with complex relations. Generated code lives in `adapters/` only — domain types are never SQLBoiler structs.

### 7q. Disable test caching with `-count=1`

Use `-count=1` to disable go test caching for tests that rely on Docker infrastructure:

```bash
go test -count=1 ./...
```

This prevents mistaking a cached success for a real pass when you change docker-compose or environment variables.

### 7r. Build tags to separate Docker-required tests

```go
// +build docker
```

Run only Docker tests:

```bash
go test -tags=docker ./...
```

The book's project was small enough to not need this separation, but larger projects benefit from running unit tests in fast feedback loops before spinning up Docker.

---

## 8. CI/CD for Integration Tests

### 8a. docker-compose base file plus CI override

Keep a single `docker-compose.yml` base file. Use an override file `docker-compose.ci.yml` for CI-specific changes:

```yaml
# docker-compose.ci.yml
services:
  trainer-http:
    image: "gcr.io/${PROJECT_ID}/trainer"
  trainer-grpc:
    image: "gcr.io/${PROJECT_ID}/trainer"
  trainings-http:
    image: "gcr.io/${PROJECT_ID}/trainings"
  users-http:
    image: "gcr.io/${PROJECT_ID}/users"
  users-grpc:
    image: "gcr.io/${PROJECT_ID}/users"

networks:
  default:
    external:
      name: cloudbuild
```

Run with both files:

```bash
docker-compose -f docker-compose.yml -f docker-compose.ci.yml up -d
```

On Google Cloud Build, all containers live inside the `cloudbuild` network — setting that as the default network lets CI test steps reach docker-compose services.

The CI uses pre-built images (not rebuilding from `Dockerfile`), so tests verify the same images that will be deployed.

### 8b. Test before deploy

In the CI pipeline, run tests after Docker images are built but before deploying. The `waitFor` key in Cloud Build allows parallelism:

```yaml
- id: docker-compose
  name: 'docker/compose:1.19.0'
  args: ['-f', 'docker-compose.yml', '-f', 'docker-compose.ci.yml', 'up', '-d']
  env:
    - 'PROJECT_ID=$PROJECT_ID'
  waitFor: [trainer-docker, trainings-docker, users-docker]

- id: trainer-tests
  name: golang
  entrypoint: ./scripts/test.sh
  args: ["trainer", ".test.ci.env"]
  waitFor: [docker-compose]

- id: trainings-tests
  name: golang
  entrypoint: ./scripts/test.sh
  args: ["trainings", ".test.ci.env"]
  waitFor: [docker-compose]

- id: e2e-tests
  name: golang
  entrypoint: ./scripts/test.sh
  args: ["common", ".e2e.ci.env"]
  waitFor: [trainer-tests, trainings-tests, users-tests]
```

### 8c. Environment variable management with bash scripts

Use separate `.env` files for each scenario (local, local component tests, CI component tests, local e2e, CI e2e). A test runner bash script loads the appropriate combination:

```bash
#!/bin/bash
set -e

readonly service="$1"
readonly env_file="$2"

cd "./internal/$service"
env $(cat "../../.env" "../../$env_file" | grep -Ev '^#' | xargs) go test -count=1 ./...
```

"Please don't try to define such a complex scenario directly in the Makefile. Make is terrible at managing environment variables." Use bash scripts — they can be linted with `shellcheck` and are easier to reason about.

The total pipeline from `git push` to production (including tests): **4 minutes** in Wild Workouts.

---

## 9. Strategic DDD (Introduction)

### 9a. The SaaS company story

The book opens Strategic DDD with a real story: a SaaS company where adding one change required changes in most services. Even with microservices, the architecture was so tightly coupled that they couldn't deploy independently. The business didn't understand why adding "one button" took two months. Stakeholders stopped trusting the development team.

The solution was not infrastructure — it was Strategic DDD patterns.

### 9b. Tactical vs Strategic — 30%/70% split

Tactical DDD patterns (what chapters 7–14 cover) give you 30% of DDD's value. The remaining 70% comes from Strategic patterns. Many teams use only tactical patterns without the strategic part — the book calls this "super horrifying."

Strategic DDD answers:
- What problem are you solving?
- Will your solution meet stakeholder and user expectations?
- How complex is the project?
- What features are not necessary?
- How to separate services to support fast development long-term?

### 9c. Why microservices alone don't fix design problems

A distributed monolith is worse than a monolith: all the coupling plus network overhead. The Accelerate research finding:

> "employing the latest whizzy microservices architecture deployed on containers is no guarantee of higher performance if you ignore these characteristics. [...] Architectural approaches that enable this strategy include the use of bounded contexts and APIs as a way to decouple large domains into smaller, more loosely coupled units."

Microservices without proper domain boundaries create services where every feature requires changes across multiple services, independent deployment is impossible, and testing in isolation is impossible.

### 9d. Bounded Contexts

Bounded Context is a Strategic DDD pattern that helps split big models into smaller, logical pieces. It is the key to achieving proper service separation. If you need to touch half the system to implement and test new functionality, your separation is wrong.

Two failure modes:
- **Wrong separation**: tightly coupled services that must change together
- **No separation**: god objects (huge objects that know too much or do too much), changes primarily affect one huge service

A great tool for discovering Bounded Contexts is Event Storming.

### 9e. Event Storming

Event Storming is a workshop during which people with questions (developers) meet people with answers (stakeholders). During the session, they rapidly explore complex business domains by building a working flow based on Domain Events (orange sticky notes).

Outcomes of a well-run Event Storming session:
- Verification that the solution meets requirements before implementation (sticky notes are cheap; code changes are expensive)
- Visibility into why adding "one button" requires significant work
- Initial view of how to split services by *responsibility* (not by entity or endpoint)
- Developed Ubiquitous Language between business and engineering

"Changing a sticky note on the board is extremely cheap. Introducing changes and verifying ideas in the developed and deployed code is hugely more expensive."

The book cites a session that stopped a project for a couple months because it revealed the core assumptions were invalid — saving months of useless development.

### 9f. Event Modeling

In 2018, Adam Dymitruk proposed the Event Modeling technique. It is heavily based on Event Storming but adds new features and puts extra emphasis on the UX part of the session. The two techniques are compatible — even if you stay with Event Storming, some Event Modeling approaches may be valuable. More at [eventmodeling.org](https://eventmodeling.org).

### 9g. Ubiquitous Language

Build a shared vocabulary between developers, operations, stakeholders, and users. Naming in code (packages, types, methods) should match the terms stakeholders use. Commands and queries named in ubiquitous language make the service navigable by anyone.

"It took me time to see how many communication issues between developers and non-developers are because of using a different language. And how painful it is."

The test: go to your business stakeholders and listen to how they call operations. If your code uses different terminology, that gap is a source of bugs and miscommunication.

---

## 10. Combining DDD, CQRS, and Clean Architecture

### 10a. Why the three patterns work best together

DDD Lite, CQRS, and Clean Architecture are most powerful when combined:
- DDD Lite keeps domain logic out of handlers and repositories
- Clean Architecture gives the domain a safe space where infrastructure cannot intrude
- CQRS makes the application layer navigable and makes dependencies trivial to mock in tests
- Repository pattern with `updateFn` closures makes transactions work cleanly across all three layers

Each pattern supports the others. The combination gives you the ability to keep constant development speed without destroying and touching existing code too much.

### 10b. The `TrainingsFirestoreRepository` — all three patterns connected

The Firestore implementation of `training.Repository` shows the patterns working together. The repository enforces authorization (`CanUserSeeTraining`) before calling `updateFn`, which is provided by the command handler, which calls domain methods:

```go
// adapters/trainings_firestore_repository.go
func (r TrainingsFirestoreRepository) UpdateTraining(
    ctx context.Context,
    trainingUUID string,
    user training.User,
    updateFn func(ctx context.Context, tr *training.Training) (*training.Training, error),
) error {
    trainingsCollection := r.trainingsCollection()

    return r.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        documentRef := trainingsCollection.Doc(trainingUUID)

        firestoreTraining, err := tx.Get(documentRef)
        if err != nil {
            return errors.Wrap(err, "unable to get actual docs")
        }

        tr, err := r.unmarshalTraining(firestoreTraining)
        if err != nil {
            return err
        }

        if err := training.CanUserSeeTraining(user, *tr); err != nil {
            return err
        }

        updatedTraining, err := updateFn(ctx, tr)
        if err != nil {
            return err
        }

        return tx.Set(documentRef, r.marshalTraining(updatedTraining))
    })
}
```

The full chain: HTTP handler (port) → command struct → `ApproveTrainingRescheduleHandler.Handle` (app) → `UpdateTraining` with `updateFn` closure (repository/adapter) → `tr.ApproveReschedule` (domain) → Firestore.

### 10c. Repository refactoring is proportional to domain improvement

Before DDD: one repository method per use case (7 methods). After DDD + CQRS: 3 methods (`AddTraining`, `GetTraining`, `UpdateTraining`). The domain entity handles all the use-case-specific logic. This is a reliable signal of correct separation.

### 10d. Domain logic is stable; infrastructure changes are safe

"It may depend on the project, but often domain logic is pretty stable after the initial development and can live unchanged for a long time. It can survive moving between services, framework changes, library changes, and API changes. Thanks to that separation, we can do all these changes in a much safer and faster way."

---

## 11. General Principles

### Quality vs speed is a false tradeoff

Skipping tests and clean architecture saves time now but costs 5x later. "We'll just rewrite the service" almost never happens cleanly. Teams that sacrifice quality for speed become afraid to make changes, which slows them down permanently.

From Accelerate research: there is no tradeoff between speed and stability. High-performing teams achieve both simultaneously. Quality is an investment, not a cost. "There is no quality vs. speed tradeoff. If you want to go fast in the long term, you need to keep high quality."

### 5 days of coding saves 1 day of planning

"With 5 days of coding, you can save 1 day of planning." The unasked questions will not magically disappear. Take time to understand the problem before implementing. Event Storming, even a short 10-minute session for a single story, helps validate assumptions before they become code.

### Prefer PostgreSQL over NoSQL for new projects

The book originally used Firestore but the repository pattern abstracts the choice. PostgreSQL is easier to operate, simpler to migrate, and avoids vendor lock-in. Abstract the database behind a repository interface so the choice can change.

### Don't pass application logic through `context.Context`

Use explicit parameters. `context.Context` is for cancellation, deadlines, and tracing — not for carrying business data or authentication. Passing values through context loses static typing, hides function inputs, and is a symptom of a function doing too much or a bad design somewhere.

### Enforce layer dependencies with static analysis

Use [go-cleanarch](https://github.com/roblaszczak/go-cleanarch) in CI to prevent adapters from importing ports, or domain from importing anything outside itself. Catches architectural violations automatically.

### Don't over-engineer simple services

"If you are creating a project that is not complex and will not be touched any time soon after 1 month of development, probably it's enough to put everything in one `main` package. Just keep in mind when this 1 month of development will become one year!"

The `users` microservice in Wild Workouts is intentionally kept simple (no Clean Architecture, no CQRS) because it has no domain logic. Apply these patterns where the domain is complex enough to justify them.
