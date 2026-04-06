# Domain-Driven Design with Golang — Rules for Claude

*Source: "Domain-Driven Design with Golang" by Matthew Boyle, Packt Publishing, 2022*

---

## 1. When to Use DDD

DDD is not a universal fit. Apply Vaughn Vernon's DDD scorecard before committing:

| Criterion | Points |
|---|---|
| Mostly simple CRUD with no business logic between input and output | 0 |
| Fewer than 30 user stories / business flows | 0 |
| 40+ user stories / business flows | 1 |
| Application is likely to grow in complexity | 1 |
| Long-lived system with non-trivial upcoming changes | 1 |
| Domain is new and no one has built it before | 1 |

**Score ≥ 7 → strong DDD candidate.** Score < 7 → consider using only selected DDD patterns rather than the full approach.

DDD requires full stakeholder commitment — it cannot be an engineering-only decision. Domain experts must be actively involved. It requires engineers to think about behaviour first, code second.

Simple CRUD services (such as a basic `users` microservice that stores and retrieves profiles) do not benefit from DDD. Forcing DDD there adds overhead without payoff.

---

## 2. Three Pillars of DDD

1. **Ubiquitous Language** — a shared vocabulary between domain experts and engineers that is reflected in code naming, package names, types, and methods.

2. **Strategic Design** — mapping out the business domain, defining bounded contexts, and choosing the right service boundaries.

3. **Tactical Design** — the implementation-level patterns: entities, value objects, aggregates, factories, repositories, and services.

---

## 3. Domains and Subdomains

A **domain** is "a sphere of knowledge, influence, or activity" (Evans). Every time you read "domain-driven design," read it as "business problem-driven design."

Subdomains are children of higher-level domains. The distinction is contextual — use "subdomain" when signalling a parent-child relationship.

Bigger companies organize teams around domains. Domain discovery is a discussion involving all stakeholders, not just engineers. DDD is not a science; there are often multiple valid decompositions.

---

## 4. Ubiquitous Language

Ubiquitous language is the overlap between domain-expert language and technical language.

Rules:
- Every type, function, method, and package name should use the language domain experts use in conversation.
- The language evolves as your team's understanding of the domain deepens.
- It must not be imposed by domain experts alone — it is a collaboration.
- If your business stakeholders use a term, your code should use the same term.
- **Keep a glossary** of terms in the team's wiki, reviewed regularly.
- Spend time with domain experts. Join their meetings, take minutes. Write down any terms you didn't understand and follow up for definitions.
- Evaluate and update the language regularly (e.g., during sprint planning).

**Concrete example:** If the business says "coffee lovers" not "customers," name the type `CoffeeLover`, not `Customer`. If they say "churn," that word belongs in your code:

```go
const (
    unknownUserType UserType = iota
    lead
    customer
    churned
    lostLead
)
```

The domain model should mirror what the business talks about. If domain experts say "a lead converts to a customer when they pick a subscription," the code should read that way:

```go
type LeadRequest struct {
    email string
}

type Lead struct {
    id string
}

type LeadCreator interface {
    CreateLead(ctx context.Context, request LeadRequest) (Lead, error)
}

type LeadConvertor interface {
    Convert(ctx context.Context, subSelection SubscriptionType) (Customer, error)
}

func (l Lead) Convert(ctx context.Context, subSelection SubscriptionType) (Customer, error) {
    // implementation
}
```

This reflects leads, converting leads, customers, and subscriptions — all ubiquitous language. Compare to the original code which used `AddUser` and `UserManager` — terms the domain experts never used.

**Failure mode example:** A business asked for "support for multiple accounts per customer," but engineers had built around a one-user-per-account assumption and used the term "user" not "customer" — missing an important invariant. The requirement got lost in translation because the ubiquitous language was never established.

**Warning:** Do not try to apply a ubiquitous language across multiple projects, teams, or an entire company. Evans advises it should only apply to a single bounded context. Trying to make a loaded term such as "customer" apply everywhere will cause the term to lose rigor.

The ubiquitous language also forms the basis for test names (see section 11).

---

## 5. Bounded Contexts

A **bounded context** defines the boundary within which a particular model and ubiquitous language apply. The same word may mean different things in different bounded contexts.

Key rules:
- A bounded context is a clean architectural boundary with its own domain model.
- Services inside a bounded context expect transactional (immediate, atomic) consistency.
- Communication *between* bounded contexts should expect eventual consistency.
- If you need to touch half the system to implement and test new functionality, your boundaries are wrong.

**Two failure modes:**
- Wrong separation: tightly coupled contexts that must change together (distributed monolith)
- No separation: god objects that know and do too much

Use **Open Host Service** (a published API) to let other bounded contexts access your context cleanly. Use **Anti-Corruption Layers** to translate between them.

When someone from a different area of the business discusses "customers," first define what a customer means in *their* bounded context — it may mean something entirely different.

---

## 6. Open Host Service and Published Language

An **Open Host Service** is the API you expose to other bounded contexts. Evans leaves implementation purposefully ambiguous — it is typically an RPC: RESTful API, gRPC, or XML API.

Example: exposing an HTTP endpoint so another team can check whether a user has an active subscription:

```go
type UserHandler interface {
    IsUserSubscriptionActive(ctx context.Context, userID string) bool
}

type UserActiveResponse struct {
    IsActive bool
}

func router(u UserHandler) {
    m := mux.NewRouter()
    m.HandleFunc("/user/{userID}/subscription/active",
        func(w http.ResponseWriter, r *http.Request) {
            uID := mux.Vars(r)["userID"]
            if uID == "" {
                w.WriteHeader(http.StatusBadRequest)
                return
            }
            isActive := u.IsUserSubscriptionActive(r.Context(), uID)
            b, err := json.Marshal(UserActiveResponse{IsActive: isActive})
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            _, _ = w.Write(b)
        }).Methods(http.MethodGet)
}
```

A **Published Language** is the format in which that API communicates. Two common choices:

### OpenAPI / Swagger

Documentation-first. Generates both client and server code. Use `oapi-codegen` for Go:

```sh
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
oapi-codegen --config=config.yml ./oapi.yaml
```

Config file:
```yaml
package: oapi
output: ./openapi.gen.go
generate:
  models: true
  client: true
```

Advantages: documentation-first — external-facing docs are always up to date; code generation supports many use cases with no extra effort. Set up a CI job that regenerates the client package whenever the spec changes, so consumer teams always get the latest version.

Downsides: no native breaking-change detection (removing a field breaks consumers without warning); less performant than gRPC.

### gRPC + protobuf

Use `buf` to generate Go client/server from `.proto` files:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
brew install bufbuild/buf/buf
buf generate
```

Example proto:
```proto
syntax = "proto3";
package user.v1;
option go_package = "example.com/testing/protos/user";

message User {
  int64 id = 1;
  string username = 2;
  string email = 3;
}

service UserService {
  rpc CreateUser (CreateUserRequest) returns (CreateUserResponse) {}
}
```

gRPC advantages: binary serialization, load balancing, health checks, bi-directional streaming, built-in auth. More complex setup than REST.

**Which to choose:** Either works well. gRPC is becoming more popular due to performance. OpenAPI is easier to retrofit to existing REST APIs.

---

## 7. Anti-Corruption Layer (ACL)

Also called the **adapter layer**. Translates models between bounded contexts or external systems. Prevents external models from corrupting your domain model.

```go
type Campaign struct {
    ID      string
    Title   string
    Goal    string
    EndDate time.Time
}

type MarketingCampaignModel struct {
    Id       string `json:"id"`
    Metadata struct {
        Name     string `json:"name"`
        Category string `json:"category"`
        EndDate  string `json:"endDate"`
    } `json:"metadata"`
}

func (m *MarketingCampaignModel) ToCampaign() (*Campaign, error) {
    if m.Id == "" {
        return nil, errors.New("campaign ID cannot be empty")
    }
    formattedDate, err := time.Parse("2006-01-02", m.Metadata.EndDate)
    if err != nil {
        return nil, errors.New("endDate was not in a parsable format")
    }
    return &Campaign{
        ID:      m.Id,
        Title:   m.Metadata.Name,
        Goal:    m.Metadata.Category,
        EndDate: formattedDate,
    }, nil
}
```

Rules:
- **Do not** swap your domain model to match an external model — that couples your model to something outside your control.
- ACLs in complex systems can be entire services (e.g., for multi-stage migrations).
- ACLs add latency and a point of failure — use them deliberately.
- The ACL is the most universally applicable DDD pattern, even in non-DDD codebases.

---

## 8. Entities

An entity is defined by its **identity**, not its attributes. Its attributes may change; its identity does not.

```go
type Auction struct {
    ID            int
    startingPrice money.Money
    sellerID      int
    createdAt     time.Time
    auctionStart  time.Time
    auctionEnd    time.Time
}
```

### Generating Entity IDs

- Avoid plain `int` for IDs — they overflow:

```go
fmt.Println(math.MaxInt)
// output: 9223372036854775807

fmt.Println(math.MaxInt + 1)
// error: cannot use math.MaxInt + 1 (untyped int constant 9223372036854775808) as int value (overflows)
```

- Prefer UUIDs (128-bit labels, effectively unique when generated per spec). Use `github.com/google/uuid`:

```go
type SomeEntity struct {
    id uuid.UUID
}

func NewSomeEntity() *SomeEntity {
    return &SomeEntity{id: uuid.New()}
}
```

- Email addresses make poor identifiers — users can change them. Tax IDs or system-generated UUIDs are better.
- PostgreSQL can generate UUIDs for you.

### Avoiding Anemic Domain Models

An **anemic model** has mostly public getters/setters and no business logic. It forces callers to implement business logic, leading to inconsistency.

**Anemic (bad)** — mostly public getters and setters, no business logic, forces callers to implement invariants themselves:

```go
type AnemicAuction struct {
    id            int
    startingPrice money.Money
    sellerID      int
    createdAt     time.Time
    auctionStart  time.Time
    auctionEnd    time.Time
}

func (a *AnemicAuction) GetID() int                          { return a.id }
func (a *AnemicAuction) StartingPrice() money.Money          { return a.startingPrice }
func (a *AnemicAuction) SetStartingPrice(p money.Money)      { a.startingPrice = p }
func (a *AnemicAuction) GetSellerID() int                    { return a.sellerID }
func (a *AnemicAuction) SetSellerID(id int)                  { a.sellerID = id }
func (a *AnemicAuction) GetCreatedAt() time.Time             { return a.createdAt }
func (a *AnemicAuction) SetCreatedAt(t time.Time)            { a.createdAt = t }
func (a *AnemicAuction) GetAuctionStart() time.Time          { return a.auctionStart }
func (a *AnemicAuction) SetAuctionStart(t time.Time)         { a.auctionStart = t }
func (a *AnemicAuction) GetAuctionEnd() time.Time            { return a.auctionEnd }
func (a *AnemicAuction) SetAuctionEnd(auctionEnd time.Time)  { a.auctionEnd = auctionEnd }
```

Consumers of `AnemicAuction` will make assumptions about attributes and implement business logic themselves in inconsistent ways.

**Rich entity (good):**
```go
func (a *AuctionRefactored) SetAuctionEnd(auctionEnd time.Time) error {
    if err := a.validateTimeZone(auctionEnd); err != nil {
        return err
    }
    a.auctionEnd = auctionEnd
    return nil
}

func (a *AuctionRefactored) validateTimeZone(t time.Time) error {
    tz, _ := t.Zone()
    if tz != time.UTC.String() {
        return errors.New("time zone must be UTC")
    }
    return nil
}
```

The domain entity enforces its own invariants. Callers cannot put it in an invalid state.

**Additional methods on a rich entity** demonstrate the pattern clearly:

```go
func (a *AuctionRefactored) GetAuctionElapsedDuration() time.Duration {
    return a.auctionStart.Sub(a.auctionEnd)
}

func (a *AuctionRefactored) GetAuctionEndTimeInUTC() time.Time {
    return a.auctionEnd
}

func (a *AuctionRefactored) GetAuctionStartTimeInUTC() time.Time {
    return a.auctionStart
}
```

Benefits: guaranteed consistent time zones, clearly communicates that UTC is required, provides a single definition of elapsed duration preventing consumer drift.

**Why not use the same model for the database?** As systems grow, you may need to store metadata (view counts, ad effectiveness, tracing IDs) that does not belong in the domain model. The domain model and persistence model should be separate.

### ORMs and Entities

Author's position: ORMs lead to unnecessary abstraction and poor database query design — a common source of performance issues.

If you use GORM or similar:
- Do not let the ORM control how you write entities — it risks anemic models.
- Keep coupling between your entity and ORM to a minimum.
- Use an adapter layer to decouple ORM and DDD entity layers.
- Never let the database schema drive the domain model.

---

## 9. Value Objects

A value object is defined by its **values**, not its identity. Two value objects with the same values are equal. They have no identity.

Rules:
- Value objects must be **immutable** — unexported fields, no setters.
- Value objects must be **comparable by value** — return structs, not pointers.
- Value objects should be **side-effect free** — functions on them return new instances.
- Use value objects as much as possible. Upgrade to entities only when you need identity.

**Key Go rule:** Return a struct (not a pointer) to get value comparison. Returning `&Point` stores the value in a new memory address — two calls with the same values produce different pointers, so `a != b` even when values are identical:

```go
// BAD — pointer comparison, always unequal:
func NewPoint(x, y int) *Point { return &Point{x: x, y: y} }

// This test FAILS:
func Test_Point(t *testing.T) {
    a := NewPoint(1, 1)
    b := NewPoint(1, 1)
    if a != b { t.Fatal("a and b were not equal") }
    // output: a and b were not equal (different memory addresses)
}

// GOOD — value comparison, equal when values match:
func NewPoint(x, y int) Point { return Point{x: x, y: y} }
// Same test now PASSES.
```

**Replaceability example:**
```go
func move(currLocation Point, direction int) Point {
    switch direction {
    case directionNorth:
        return NewPoint(currLocation.x, currLocation.y+1)
    // ...
    }
    return currLocation
}
```

Each move returns a completely new `Point` — the old one is replaced. The `move` function is side-effect free.

**Benefits of immutability and side-effect-free functions:**
- Easier to reason about
- Easier to write unit tests for (predictable inputs → predictable outputs)
- Supports multiple inputs without unexpected state changes
- Aids long-term system maintenance

**Decision guide — value object if you can answer yes to all:**
- Can I treat this as immutable?
- Does it measure, quantify, or describe a domain concept?
- Can it be compared by values alone?

**Practical strategy:** Try to make everything a value object to start with. When it no longer fits the use case, upgrade it to an entity.

---

## 10. Aggregates

An aggregate is a cluster of domain objects (entities and value objects) treated as a single unit for certain purposes. The aggregate root is the entry point and holds the identity.

```go
type WalletItem interface {
    GetBalance() (money.Money, error)
}

type Wallet struct {
    id          uuid.UUID
    ownerID     uuid.UUID
    walletItems []WalletItem
}

func (w Wallet) GetWalletBalance() (*money.Money, error) {
    var bal *money.Money
    for _, v := range w.walletItems {
        itemBal, err := v.GetBalance()
        if err != nil {
            return nil, errors.New("failed to get balance")
        }
        bal, err = bal.Add(&itemBal)
        if err != nil {
            return nil, errors.New("failed to increment balance")
        }
    }
    return bal, nil
}
```

### Rules for Aggregates

- The aggregate acts as a **transactional consistency boundary**: load, save, edit, and delete should happen to all objects within it or not at all.
- Ideally, **modify only one aggregate per transaction**. If you need more, your model is probably wrong.
- Aim for **small aggregates** — smaller aggregates improve scalability, performance, and transaction success rates.
- Keep only what belongs: don't add marketing opt-in to an order aggregate just because the UI has a checkbox for it.
- **Beyond a single bounded context**, expect and aim for eventual consistency.

### Discovering Aggregates

Find **invariants** (business rules that must always be true) first. Then group the domain objects that must be consistent together to uphold those invariants.

**Wrong: large aggregate**
```go
type Order struct {
    items          []item
    taxAmount      money.Money
    discount       money.Money
    paymentCardID  uuid.UUID
    customerID     uuid.UUID
    marketingOptIn bool  // ← wrong: different bounded context, different consistency requirement
}
```

**Right: focused aggregate**
```go
type Order struct {
    items         []item
    taxAmount     money.Money
    discount      money.Money
    paymentCardID uuid.UUID
    customerID    uuid.UUID
}
```

Aggregates are not the same as slices/maps/arrays. They are DDD concepts with business logic, transactions, and domain rules.

---

## 11. Factory Pattern

A factory is a function (or struct) whose primary responsibility is creating other domain objects. It enforces invariants at creation time.

```go
func BuildCar(carType string) (Car, error) {
    switch carType {
    case "bmw":
        return BMW{heatedSeatSubscriptionEnabled: true}, nil
    case "tesla":
        return Tesla{autoPilotEnabled: true}, nil
    default:
        return nil, errors.New("unknown car type")
    }
}
```

**Enforcing business invariants at creation:**
```go
func CreateBooking(from, to time.Time, hairDresserID uuid.UUID) (*Booking, error) {
    closingTime, _ := time.Parse(time.Kitchen, "17:00pm")
    if from.After(closingTime) {
        return nil, errors.New("no appointments after closing time")
    }
    return &Booking{
        hairDresserID: hairDresserID,
        id:            uuid.New(),
        from:          from,
        to:            to,
    }, nil
}
```

Rules:
- Factory functions should standardize complex struct creation and provide encapsulation.
- Factories enforce business invariants at the time of object creation.
- For entities: decide whether the factory generates the ID or receives it as a parameter. Prefer letting the factory generate it unless there is a good reason not to.
- Factory functions should set sensible defaults on the objects they create.

---

## 12. Repository Pattern

A repository centralizes all data access logic. It decouples the domain from a specific database technology.

**Interface in the domain package:**
```go
type BookingRepository interface {
    SaveBooking(ctx context.Context, booking Booking) error
    DeleteBooking(ctx context.Context, booking Booking) error
}
```

**Postgres implementation (separate from domain):**
```go
type PostgresRepository struct {
    connPool *pgx.Conn
}

func NewPostgresRepository(ctx context.Context, dbConnString string) (*PostgresRepository, error) {
    conn, err := pgx.Connect(ctx, dbConnString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to db: %w", err)
    }
    defer conn.Close(ctx)
    return &PostgresRepository{connPool: conn}, nil
}

func (p PostgresRepository) SaveBooking(ctx context.Context, booking Booking) error {
    _, err := p.connPool.Exec(
        ctx,
        "INSERT into bookings (id, from, to, hair_dresser_id) VALUES ($1,$2,$3,$4)",
        booking.id.String(),
        booking.from.String(),
        booking.to.String(),
        booking.hairDresserID.String(),
    )
    if err != nil {
        return fmt.Errorf("failed to SaveBooking: %w", err)
    }
    return nil
}
```

### Key Rules

- **One repository per aggregate**, not one per database table. A repository can write to multiple tables.
- The repository interface lives in the domain package. Implementations live in infrastructure/adapter packages.
- No domain logic in repositories — that belongs in the service or domain layer.
- Define the interface first; implementations can be provided later. This allows teams to develop in parallel.
- Decouple persistence models from domain models using adapter structs:

```go
type mongoPurchase struct {
    id                 uuid.UUID
    store              store.Store
    productsToPurchase []coffeeco.Product
    total              money.Money
    paymentMeans       payment.Means
    timeOfPurchase     time.Time
    cardToken          *string
}

func toMongoPurchase(p Purchase) mongoPurchase {
    return mongoPurchase{
        id:                 p.id,
        store:              p.Store,
        productsToPurchase: p.ProductsToPurchase,
        total:              p.total,
        paymentMeans:       p.PaymentMeans,
        timeOfPurchase:     p.timeOfPurchase,
        cardToken:          p.cardToken,
    }
}
```

---

## 13. Services

DDD uses three types of services. Know which is which.

### Domain Services

Stateless operations within a domain that don't fit naturally into an entity or value object.

Use a domain service when:
- You are performing significant business logic within one domain
- You are transforming one domain object into another
- You are calculating a value using properties from two or more domain objects

```go
// BAD: ShoppingCart references Product and contains cross-entity logic
func (s *ShoppingCart) AddToCart(p Product) bool {
    if p.CanBeBought() { ... }
}

// GOOD: domain service handles cross-entity logic
type ProductOrderService struct{}

func (s ProductOrderService) AddProductToCart(
    ctx context.Context,
    cart *ShoppingCart,
    product Product,
) error {
    if !product.CanBeBought() {
        return errors.New("product cannot be bought")
    }
    return cart.AddItem(product)
}
```

Domain services must use ubiquitous language from within the bounded context.

### Application Services

Orchestrate domain services and repositories. Handle authorization, compose multiple domain services, power UIs. Application services should be **thin** — all business logic is pushed down into domain services and domain objects.

The application service also handles security concerns (authentication, authorization):

```go
type BookingDomainService interface {
    CreateBooking(ctx context.Context, booking Booking) error
}

type BookingAppService struct {
    bookingRepo          BookingRepository
    bookingDomainService BookingDomainService
    emailService         EmailSender
}

func (b *BookingAppService) CreateBooking(ctx context.Context, booking Booking) error {
    // Security: extract and validate caller identity from context
    u, ok := ctx.Value(accountCtxKey).(*chapter2.Customer)
    if !ok {
        return errors.New("invalid customer")
    }
    if u.UserID() != booking.userID.String() {
        return errors.New("cannot create booking for other users")
    }
    if err := b.bookingDomainService.CreateBooking(ctx, booking); err != nil {
        return fmt.Errorf("could not create booking: %w", err)
    }
    if err := b.bookingRepo.SaveBooking(ctx, booking); err != nil {
        return fmt.Errorf("could not save booking: %w", err)
    }
    if err := b.emailService.SendEmail(ctx, ...); err != nil {
        // handle
    }
    return nil
}
```

Another common use: powering a UI that needs to compose many different domain services to show a single screen.

### Infrastructure Services

Third-party integrations (payment, email, analytics) that are not part of your domain. Always hide behind an interface:

```go
type EmailSender interface {
    SendEmail(ctx context.Context, to string, title string, body string) error
}

type MailChimp struct {
    apiKey     string
    from       string
    httpClient http.Client
}

func (m MailChimp) SendEmail(ctx context.Context, to string, title string, body string) error {
    // ... implementation
}
```

This decouples your domain from specific providers. If Stripe becomes too expensive, only the `StripeService` adapter needs to change — the domain and application service are untouched.

---

## 14. Project Structure in Go

- Put all domain code inside an `internal/` folder. Go enforces that nothing outside the module can import `internal/` packages.
- Organize by domain, not by technical layer:

```
coffeeco/
  internal/
    coffeelover.go        ← root domain entity
    product.go            ← shared value object
    store/
      store.go            ← Store entity
      repository.go       ← Store repository interface + implementation
    loyalty/
      coffeebux.go        ← CoffeeBux aggregate
    purchase/
      purchase.go         ← Purchase entity + Service
      repository.go       ← Purchase repository interface + implementation
    payment/
      means.go            ← Payment type aliases/constants
      stripe.go           ← Stripe infrastructure service
  cmd/
    main.go               ← wiring only
```

- Infrastructure services (Stripe, MailChimp) go inside the relevant domain package.
- Repository implementations go inside the domain package alongside the interface (or in a subpackage for large implementations).
- `main.go` is for wiring dependencies only; no business logic.

---

## 15. Defensive Programming Patterns

Always validate constructor arguments:

```go
func NewService(availability AvailabilityGetter) (*Service, error) {
    if availability == nil {
        return nil, errors.New("availability must not be nil")
    }
    return &Service{availability: availability}, nil
}
```

Always validate domain invariants in domain methods:

```go
func (p *Purchase) validateAndEnrich() error {
    if len(p.ProductsToPurchase) == 0 {
        return errors.New("purchase must consist of at least one product")
    }
    p.total = *money.New(0, "USD")
    for _, v := range p.ProductsToPurchase {
        newTotal, _ := p.total.Add(&v.BasePrice)
        p.total = *newTotal
    }
    if p.total.IsZero() {
        return errors.New("likely mistake; purchase should never be 0. Please validate")
    }
    p.id = uuid.New()
    p.timeOfPurchase = time.Now()
    return nil
}
```

Push logic down into domain objects as much as possible. Keep services thin.

**Rich domain methods on aggregates** (CoffeeBux loyalty scheme):

```go
// AddStamp adds a virtual loyalty stamp after each purchase.
// When the counter reaches 1, the customer has earned a free drink.
func (c *CoffeeBux) AddStamp() {
    if c.RemainingDrinkPurchasesUntilFreeDrink == 1 {
        c.RemainingDrinkPurchasesUntilFreeDrink = 10
        c.FreeDrinksAvailable++
    } else {
        c.RemainingDrinkPurchasesUntilFreeDrink--
    }
}

// Pay redeems free drinks from the loyalty card.
func (c *CoffeeBux) Pay(ctx context.Context, purchases []Purchase) error {
    lp := len(purchases)
    if lp == 0 {
        return errors.New("nothing to buy")
    }
    if c.FreeDrinksAvailable < lp {
        return fmt.Errorf("not enough coffeeBux to cover purchase. Have %d, need %d",
            c.FreeDrinksAvailable, lp)
    }
    c.FreeDrinksAvailable -= lp
    return nil
}
```

Note the defensive validation: checks that the purchase list is not empty and that there are enough free drinks. An assumption is made that partial redemption is not allowed — this must be validated with domain experts.

**Sentinel errors** let callers distinguish between "no result" and a real error:

```go
var ErrNoDiscount = errors.New("no discount for store")

func (m MongoRepository) GetStoreDiscount(ctx context.Context, storeID uuid.UUID) (float32, error) {
    var discount float32
    err := m.storeDiscounts.FindOne(ctx, bson.D{{"store_id", storeID.String()}}).Decode(&discount)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return 0, ErrNoDiscount  // caller can handle "no discount" separately from real errors
        }
        return 0, fmt.Errorf("failed to find discount for store: %w", err)
    }
    return discount, nil
}
```

**Extract method for readability** — keep `CompletePurchase` thin by extracting logic into a helper:

```go
func (s Service) CompletePurchase(ctx context.Context, storeID uuid.UUID, purchase *Purchase, coffeeBuxCard *loyalty.CoffeeBux) error {
    if err := purchase.validateAndEnrich(); err != nil {
        return err
    }
    if err := s.calculateStoreSpecificDiscount(ctx, storeID, purchase); err != nil {
        return err
    }
    // payment switch...
    if err := s.purchaseRepo.Store(ctx, *purchase); err != nil {
        return errors.New("failed to store purchase")
    }
    if coffeeBuxCard != nil {
        coffeeBuxCard.AddStamp()
    }
    return nil
}

func (s *Service) calculateStoreSpecificDiscount(ctx context.Context, storeID uuid.UUID, purchase *Purchase) error {
    discount, err := s.storeService.GetStoreSpecificDiscount(ctx, storeID)
    if err != nil && err != store.ErrNoDiscount {
        return fmt.Errorf("failed to get discount: %w", err)
    }
    if discount > 0 {
        purchase.total = *purchase.total.Multiply(int64(100 - discount))
    }
    return nil
}
```

Note: `coffeeBuxCard` is a pointer because a customer is under no obligation to present a loyalty card — it is legitimately nil.

---

## 16. Microservices and DDD

Microservices are an organizational decision as much as a technical one. They are not a silver bullet.

**Adopt when:**
- Teams need to move fast independently
- Different parts of the system have different scaling requirements
- You want freedom to choose different technologies per service

**Answer these questions before adopting microservices:**
- Do we have expertise to run a distributed system? Strategy to hire/train?
- Do we have tooling to monitor a distributed system? Budget and time?
- Which platform (Kubernetes)? Team comfort level?
- Team comfort with CI/CD pipelines?
- Will leadership invest money and time?

**Be cautious of:**
- Distributed systems require more expertise: Kubernetes, networking, latency
- Testing cross-service journeys is harder
- You need good observability tooling *before* splitting

**DDD patterns that help with microservices:**
- Bounded contexts define natural service boundaries
- Anti-corruption layers translate between services
- Open Host Services (APIs) let services communicate without tight coupling
- Domain events enable eventual consistency across service boundaries

### Resilience in Microservices

Expect failure. Use a retryable HTTP client for unreliable dependencies:

```go
import "github.com/hashicorp/go-retryablehttp"

func main() {
    c := retryablehttp.NewClient()
    c.RetryMax = 10
    partnerAdaptor, err := recommendation.NewPartnerShipAdaptor(
        c.StandardClient(),
        "http://localhost:3031",
    )
    // ...
}
```

Define interfaces for external dependencies and write adapters that satisfy them. This decouples your domain from the external API's shape:

```go
// Domain interface — uses your bounded context's language:
type AvailabilityGetter interface {
    GetAvailability(ctx context.Context, tripStart, tripEnd time.Time, location string) ([]Option, error)
}

// Adapter — knows about the external API's shape:
type PartnershipAdaptor struct {
    client *http.Client
    url    string
}

func (p PartnershipAdaptor) GetAvailability(...) ([]Option, error) {
    // HTTP call, response decoding, model translation
}
```

When the external API changes or is replaced, only the adapter changes.

### Transport Layer Separation

Decouple transport-specific concerns from domain code. A `transport` package creates the HTTP router but imports domain handlers — the domain knows nothing about HTTP:

```go
package transport

func NewMux(recHandler recommendation.Handler) *mux.Router {
    m := mux.NewRouter()
    m.HandleFunc("/recommendation", recHandler.GetRecommendation).Methods(http.MethodGet)
    return m
}
```

### Startup Validation Pattern

If a required dependency can't be created, fail fast with `log.Fatal` in `main.go`. There is no point proceeding if pre-requisite conditions are not met:

```go
svc, err := recommendation.NewService(partnerAdaptor)
if err != nil {
    log.Fatal("failed to create a service: ", err)
}
handler, err := recommendation.NewHandler(*svc)
if err != nil {
    log.Fatal("failed to create a handler: ", err)
}
```

---

## 17. Distributed System Characteristics

A distributed system is various computing components spread out over a network that coordinate to complete tasks. Any system you build on a cloud provider (AWS, Cloudflare, DigitalOcean) is a distributed system.

Characteristics:

| Property | Meaning |
|---|---|
| **Scalable** | Can grow as workloads increase; scale up/down to match traffic and optimize cost |
| **Fault-tolerant** | If one piece fails, the rest continues (a payment service crashing should not take down the whole site) |
| **Transparent** | Appears as a single unit to end users; the underlying implementation is hidden |
| **Concurrent** | Multiple activities can happen at the same time |
| **Heterogeneous** | Can mix servers, languages, and paradigms (Windows/Linux, Go/Python, event-driven/synchronous) |
| **Replicated** | Information is often replicated for fault tolerance (primary Postgres + read-only secondary) |

Trade-offs are unavoidable — the CAP theorem (§20) captures the core constraint.

---

## 18. CQRS

CQRS (Command Query Responsibility Segregation) separates reads (queries) from writes (commands). Each method is one or the other — never both.

**Origin:** Bertrand Meyer (creator of the Eiffel programming language) stated: *"Every method should be either a command that performs an action, or a query that answers a question — no method should do both. Asking a question should not change the answer."*

**Motivation:** In traditional systems, the same data model and repository are used for reads and writes. As complexity grows, this becomes unmanageable. Analytics systems write far more than they read; recommendation systems read far more than they write. CQRS lets each side scale independently.

CQRS is not required for DDD — but the complex systems that benefit from DDD often benefit from CQRS too.

**Rules adapted for Go:**
- If a method modifies the state of the receiver struct or database, it is a **command** and should return `error` or `nil`.
- If a method returns a value, it should **not** modify the database or its receiver struct (it is a **query**).

**Do not enforce CQRS with generic interfaces:**
```go
// BAD — loses type safety and clarity:
type Commander interface {
    Command(ctx context.Context, args ...interface{}) error
}
```

Use concrete, well-named command and query structs instead:
```go
// GOOD — specific, named, type-safe:
func (h ApproveRescheduleHandler) Handle(ctx context.Context, cmd ApproveReschedule) error { ... }
func (h AvailableHoursHandler) Handle(ctx context.Context, q AvailableHours) ([]HourDTO, error) { ... }
```

CQRS works best in distributed/event-driven systems. Commands are a great way to model domain event emission (e.g., writing to Kafka). In simple monoliths, it adds complexity without proportional benefit.

---

## 19. Event-Driven Architecture (EDA) and Domain Events

In EDA, systems produce, detect, and respond to events representing significant state changes.

**Domain events** represent business-meaningful state changes:
- `user.loggedIn`
- `purchase.failed`
- `order.created`

**Event structure** has two parts:
- **Event header**: meta-information — timestamp of emission, the system that emitted it, a UID for the specific message
- **Event body**: information about the state that changed

```json
{
  "event_type": "user.logged_in",
  "user_id": "135649039"
}
```

Popular formats for defining message schemas: JSON, Protobuf, Apache Avro, Cap'n Proto.

Rules:
- Domain events are output by one microservice and ingested by another.
- A domain event may be meaningful in one bounded context and irrelevant in another — this is expected and encouraged.
- Transform incoming events into your domain model's shape before processing.
- Individual domain events may represent only part of a longer-running task; chain them into **pipelines**. Pipelining is powerful — other systems can subscribe at any pipeline step, and new business requirements can be added easily without changing existing consumers.

### Dealing with Distributed Failures

**Two-Phase Commit (2PC):** Ask each subsystem if it can promise to do the work (preparation phase), then tell it to do it (completion phase). Disadvantage: blocking protocol — loses concurrency in best case, deadlock in worst case.

**Saga Pattern:** For each action, define a compensating action. If part of the transaction fails, execute compensating actions to roll back. Achieves eventual consistency without blocking locks.

```
Create Order → Reserve Stock → Charge Card → Send Confirmation
    ↓               ↓              ↓
Cancel Order   Release Stock  Refund Card   (compensating actions)
```

Go implementation:

```go
type Saga interface {
    Execute(ctx context.Context) error
    Rollback(ctx context.Context) error
}

type SagaManager struct {
    actions []Saga
}

func (s SagaManager) Handle(ctx context.Context) {
    for i, action := range s.actions {
        if err := action.Execute(ctx); err != nil {
            // Roll back all previously executed actions
            for j := 0; j <= i; j++ {
                if err := s.actions[j].Rollback(ctx); err != nil {
                    // Compensating action failed — emit an event to a message
                    // bus so it can be retried by consumers at their own pace.
                }
            }
            return
        }
    }
}
```

If a compensating action itself fails, combine the saga with EDA — emit an event for the compensation so consumer services can retry it asynchronously.

---

## 20. CAP Theorem — Choosing Databases

CAP theorem: distributed systems can guarantee at most two of Consistency, Availability, and Partition Tolerance.

- **Consistency**: Every read receives the most up-to-date information or an error.
- **Availability**: Every request receives a non-error response, but it may not be the most recent data.
- **Partition Tolerance**: The system continues to operate despite network failures. In the event of a partition, the designer must choose: cancel the operation (preserve consistency, lose availability) or proceed (preserve availability, risk inconsistency).

| Database | Type | Notes |
|---|---|---|
| MongoDB | CP | Single primary node for all writes; secondaries replicate. If the primary goes down, a secondary is promoted — the database is **unavailable during this window**. |
| Cassandra | AP | No primary node; you can write to any node. Uses consistent hashing — keys distributed across an abstract hash ring. Claims to survive complete loss of a data center (no SPOF). Prioritizes availability over strict consistency. |
| PostgreSQL | CA | Consistent and available; limited partition tolerance in distributed deployments. |

Use the CAP theorem to choose a database based on your domain's consistency requirements. If your domain requires strict consistency (financial transactions), choose CP. If availability is paramount (user-facing reads), choose AP.

---

## 21. TDD with DDD

TDD and DDD are complementary. TDD's Red-Green-Refactor cycle maps naturally to DDD's requirement to model domain behaviour.

**The TDD cycle:**
1. **Write a failing test** — name it exactly from the acceptance criteria. Use `t.FailNow()` to prevent empty tests from passing.
2. **Run it** — confirm it fails (proves the test harness works and the behaviour isn't already there).
3. **Write minimum code to pass** — spaghetti is fine at this stage; pass the test first.
4. **Run all tests** — new and old must pass.
5. **Refactor** — improve the code; re-run tests after each change.

**Test naming as documentation:**
```go
func Test_CookiePurchases(t *testing.T) {
    t.Run(`Given a user tries to purchase a cookie and we have them in stock,
        when they tap their card, they get charged and then receive an email receipt.`,
        func(t *testing.T) {
            t.FailNow() // mark incomplete until implemented
        })
}
```

Use `t.FailNow()` instead of leaving the body empty. In Go, an empty test passes immediately — that breaks the "test must fail first" rule and can silently hide gaps in test coverage.

Test names taken directly from acceptance criteria serve as living documentation. Combined with DDD's ubiquitous language, they are readable by domain experts.

### Full Worked Example: Cookie Purchase Service

The book builds `CookieService` end-to-end using TDD. This shows the complete cycle across all five acceptance criteria.

**Interfaces first** (before any implementation):

```go
type EmailSender interface {
    SendEmailReceipt(ctx context.Context, emailAddress string) error
}
type CardCharger interface {
    ChargeCard(ctx context.Context, cardToken string, amountInCents int) error
}
type CookieStockChecker interface {
    AmountInStock(ctx context.Context) int
}

type CookieService struct {
    emailSender  EmailSender
    cardCharger  CardCharger
    stockChecker CookieStockChecker
}
```

**Test 1 — Happy path (charge card, send email):**

```go
t.Run(`Given a user tries to purchase a cookie and we have them in stock,
    when they tap their card, they get charged and receive an email receipt.`,
    func(t *testing.T) {
        ctrl := gomock.NewController(t)
        e := mocks.NewMockEmailSender(ctrl)
        c := mocks.NewMockCardCharger(ctrl)
        s := mocks.NewMockCookieStockChecker(ctrl)

        cookiesToBuy := 5
        totalExpectedCost := 250  // 5 × 50 cents

        cs, _ := NewCookieService(e, c, s)
        gomock.InOrder(
            s.EXPECT().AmountInStock(ctx).Times(1).Return(cookiesToBuy),
            c.EXPECT().ChargeCard(ctx, cardToken, totalExpectedCost).Times(1).Return(nil),
            e.EXPECT().SendEmailReceipt(ctx, email).Times(1).Return(nil),
        )
        if err := cs.PurchaseCookies(ctx, cookiesToBuy, cardToken, email); err != nil {
            t.Fatalf("expected no error but got %v", err)
        }
    })
```

**Test 2 — No stock:**

```go
t.Run(`Given a user tries to purchase a cookie and we don't have any in stock,
    we return an error to the cashier.`,
    func(t *testing.T) {
        // Only AmountInStock is called — no charge, no email
        gomock.InOrder(
            s.EXPECT().AmountInStock(ctx).Times(1).Return(0),
        )
        err := cs.PurchaseCookies(ctx, 5, cardToken, email)
        if err == nil { t.Fatal("expected an error but got none") }
    })
```

Minimal code to pass: `if cookiesInStock == 0 { return errors.New("no cookies in stock sorry :(") }`

**Test 3 — Card declined:**

```go
t.Run(`Given a user tries to purchase a cookie in stock but their card gets declined,
    we return an error so the cashier can ban the customer.`,
    func(t *testing.T) {
        gomock.InOrder(
            s.EXPECT().AmountInStock(ctx).Times(1).Return(5),
            c.EXPECT().ChargeCard(ctx, cardToken, 250).Times(1).Return(errors.New("declined")),
        )
        err := cs.PurchaseCookies(ctx, 5, cardToken, email)
        if err == nil || err.Error() != "your card was declined, you are banned!" {
            t.Fatalf("unexpected error: %v", err)
        }
    })
```

**Test 4 — Email fails but transaction succeeds:**

```go
t.Run(`Given card is charged successfully but email fails,
    we return a message but the transaction is still considered done.`,
    func(t *testing.T) {
        gomock.InOrder(
            s.EXPECT().AmountInStock(ctx).Times(1).Return(5),
            c.EXPECT().ChargeCard(ctx, cardToken, 250).Times(1).Return(nil),
            e.EXPECT().SendEmailReceipt(ctx, email).Times(1).Return(errors.New("smtp down")),
        )
        err := cs.PurchaseCookies(ctx, 5, cardToken, email)
        if err == nil || err.Error() != "we are sorry but the email receipt did not send" {
            t.Fatalf("unexpected error: %v", err)
        }
    })
```

**Test 5 — More cookies requested than in stock, charge only for available:**

```go
t.Run(`Given someone wants to purchase more cookies than in stock,
    we only charge them for the ones we do have.`,
    func(t *testing.T) {
        inStock := 3
        totalExpectedCost := 150  // 3 × 50 cents, not 5 × 50

        gomock.InOrder(
            s.EXPECT().AmountInStock(ctx).Times(1).Return(inStock),
            c.EXPECT().ChargeCard(ctx, cardToken, totalExpectedCost).Times(1).Return(nil),
            e.EXPECT().SendEmailReceipt(ctx, email).Times(1).Return(nil),
        )
        if err := cs.PurchaseCookies(ctx, 5, cardToken, email); err != nil {
            t.Fatalf("expected no error but got %v", err)
        }
    })
```

Minimal fix: `if amountOfCookiesToPurchase > cookiesInStock { amountOfCookiesToPurchase = cookiesInStock }`

**Final implementation after all tests pass:**

```go
func (c *CookieService) PurchaseCookies(ctx context.Context, amount int, cardToken, email string) error {
    priceOfCookie := 50

    inStock := c.stockChecker.AmountInStock(ctx)
    if inStock == 0 {
        return errors.New("no cookies in stock sorry :(")
    }
    if amount > inStock {
        amount = inStock
    }
    if err := c.cardCharger.ChargeCard(ctx, cardToken, priceOfCookie*amount); err != nil {
        return errors.New("your card was declined, you are banned!")
    }
    if err := c.emailSender.SendEmailReceipt(ctx, email); err != nil {
        return errors.New("we are sorry but the email receipt did not send")
    }
    return nil
}
```

Note how the function signature evolved: `cardToken` and `email` were originally hardcoded placeholders. The domain expert Q&A revealed they come from the card machine — so they became parameters.

### Driving Design Through Domain Expert Q&A

During TDD, TODO comments mark open questions. Resolve them with domain experts before the refactor step:

```go
priceOfCookie := 5   // TODO: ask how much cookies cost
// TODO: where do I get cardToken from?
// TODO: what do I do if amountOfCookiesToPurchase > cookiesInStock?
```

Each answer drives a new test case and then a minimal code change. This is where TDD and DDD are most explicitly complementary — the test cycle forces you to clarify domain behaviour before writing it.

### On Table-Driven Tests

The book's explicit position: *"I am not a big proponent of table-driven tests, which are popular in the Go community; I feel they prioritize speed for the writer of the code rather than for the future reader. Code is written once but read many times, so we should always prioritize the reader."*

Each test should stand alone with all information needed to understand it, making tests the best documentation the codebase can have. Don't DRY up test boilerplate at the expense of clarity.

### Coverage

```sh
go test ./... --cover
```

### Mocking with gomock

Use `gomock` to generate interface mocks for unit testing:

```go
// gen.go at project root:
//go:generate mockgen -package mocks -destination chapter8/mocks/cookies.go \
//   github.com/example/chapter8 CookieStockChecker,CardCharger,EmailSender
```

```sh
go generate ./...
```

Usage in tests:
```go
ctrl := gomock.NewController(t)
e := mocks.NewMockEmailSender(ctrl)
c := mocks.NewMockCardCharger(ctrl)
s := mocks.NewMockCookieStockChecker(ctrl)

gomock.InOrder(
    s.EXPECT().AmountInStock(ctx).Times(1).Return(5),
    c.EXPECT().ChargeCard(ctx, "token", 250).Times(1).Return(nil),
    e.EXPECT().SendEmailReceipt(ctx, "some@email.com").Times(1).Return(nil),
)
```

Mock interfaces let you test specific failure paths without real infrastructure. Interfaces also enforce decoupling: changing from Google to AWS email only changes the adapter, not the test.

**Black-box testing:** Declare test packages as `package xxx_test` (not `package xxx`). This forces you to test only exported behaviour, making tests less brittle.

---

## 22. BDD with DDD

BDD extends TDD with a higher-level DSL readable by domain experts. Use **Gherkin** syntax:

```gherkin
Feature: checkout Integration
Scenario: Successfully Capture a payment
  Given I am a customer
  When I purchase a cookie for 50 cents
  Then my card should be charged 50 cents and an email receipt is sent
```

In Go, use `go-bdd/gobdd` (`github.com/go-bdd/gobdd`). Create a `features/` folder with `.feature` files. Step functions map to Gherkin text and pass results through a context:

```go
func add(t gobdd.StepTest, ctx gobdd.Context, first, second int) {
    res := first + second
    ctx.Set("result", res)
}

func check(t gobdd.StepTest, ctx gobdd.Context, sum int) {
    received, err := ctx.GetInt("result")
    if err != nil {
        t.Fatal(err)
        return
    }
    if sum != received {
        t.Fatalf("expected %d but got %d", sum, received)
    }
}

func TestScenarios(t *testing.T) {
    suite := gobdd.NewSuite(t)
    suite.AddStep(`I add (\d+) and (\d+)`, add)
    suite.AddStep(`the result should equal (\d+)`, check)
    suite.Run()
}
```

**BDD trade-offs:**
- BDD tests push significant complexity down into test scaffolding — expressing complex scenarios requires a lot of scaffolding.
- It is worthwhile if domain experts actively co-write acceptance criteria in Gherkin.
- If domain experts are absent or uninvested, unit tests are faster and more engaging for engineering teams.
- Much like DDD itself, BDD is a multidisciplinary team investment. Ensure buy-in from all stakeholders before committing.
- Some teams work with domain experts to write acceptance criteria in Gherkin format in their ticketing system — those criteria can then directly become tests.

---

## 23. Applying DDD to Existing Codebases

You don't need to adopt DDD all at once. Incremental approach:

1. **Build a relationship with domain experts first.** Start developing a ubiquitous language. Reflect it in naming as you touch code.

2. **Apply the ACL pattern to external dependencies.** Decouple from specific APIs (Stripe, Mailchimp, AWS). This pays off immediately and is low risk.

3. **Introduce repositories.** Even in existing code, centralizing data access improves maintainability and makes database migrations safer.

4. **Move business logic into domain objects.** Identify anemic models (public getters/setters, no invariants enforced) and gradually enrich them.

Full DDD migration (repositories, aggregates, bounded contexts) may be too large a refactor. Start small; the language and decoupling improvements provide value immediately.

---

## 24. Message Buses

A message bus decouples services by providing a common transport for events and commands. The important thing is picking the correct tool for the use case — not all message buses guarantee ordering or queue-like semantics.

### Kafka

Created at LinkedIn, open-sourced by the Apache Software Foundation. Scales to millions of requests per second. Used at internet scale (Microsoft, Cloudflare).

**Architecture:**
- **Broker**: stores messages sent to topics; topics split into partitions
- **Producers**: connect to Kafka to send messages; specify topic and partition
- **Consumers**: subscribe to topics/partitions to read messages; multiple instances form **consumer groups** that work together to consume all messages from a topic
- Services can be both producers and consumers (consume from one topic, process, produce to another)

**Challenges:** Requires deep knowledge; easy mistakes can have severe consequences (e.g., wrong partitioning strategy causes out-of-order delivery). Monitoring is difficult; running your own cluster is operationally demanding.

### RabbitMQ

Open source queuing system based on the AMQP protocol. Easy to get started; conceptually simple.

**Architecture:**
- **Publisher** sends messages to an **exchange**
- The exchange forwards messages to a **queue** based on the **routing key**
- A **consumer application** reads from the queue; once acknowledged, the message is consumed and never received again
- Built-in admin UI for visibility

**Limitations:** Does not scale as well as Kafka. Offers a subset of Kafka's features. Companies typically migrate from RabbitMQ to Kafka as they scale.

### NATS

Neural Autonomic Transport System (NATS) — open source, written in Go. The Go codebase is very readable; it's a good option for learning how streaming systems work.

**Features:**
- Publish messages to **subjects**, consumed by **subscribers**
- **Wildcard matching on subjects** (distinctive feature)
- **At-most-once delivery** — a message might never arrive at all; in return: lightweight, fast, simple to run
- Commonly used for IoT use cases

Choose NATS when simplicity and speed matter more than guaranteed delivery.

---

## 25. Money in Domain Models

Never use floats for money. Use a dedicated money library:

```go
import "github.com/Rhymond/go-money"

type Product struct {
    ItemName  string
    BasePrice money.Money
}
```

`go-money` handles arithmetic, currency conversion, and comparison correctly. This is a domain invariant: representing money incorrectly is a business bug.

---

## 26. Summary: DDD Pattern Cheat Sheet

| Pattern | Purpose | Go hint |
|---|---|---|
| Ubiquitous Language | Shared vocabulary in code and conversation | Name types/functions/packages with business terms |
| Bounded Context | Define model and language scope | Use `internal/` packages per domain |
| Entity | Identity-based object | UUID field, unexported attrs, rich methods |
| Value Object | Value-based object | No pointer return, unexported fields, no setters |
| Aggregate | Transaction boundary | One repository per aggregate root |
| Factory | Enforce invariants at creation | `func NewXxx(...) (*Xxx, error)` |
| Repository | Decouple domain from data store | Interface in domain package, impl in infra |
| Domain Service | Cross-entity business logic | Stateless, uses ubiquitous language |
| Application Service | Orchestration, authorization | Thin; composes domain services + repo |
| Infrastructure Service | Third-party integrations | Always behind an interface |
| ACL | Translate external models | `func (m ExternalModel) ToDomainModel() (*DomainModel, error)` |
| Open Host Service | Published API for other contexts | OpenAPI or gRPC |
| CQRS | Separate reads from writes | Commands return `error`; queries return data, don't mutate |
| Domain Events | Cross-context state changes | Emitted by one service, consumed by others |
| Saga | Distributed consistency | Each action has a compensating rollback action |
| Distributed System | Scalable, fault-tolerant, transparent system | Scalable, fault-tolerant, transparent, concurrent, heterogeneous, replicated |
| Kafka | High-throughput event streaming | Brokers/producers/consumers; consumer groups; complex but scales to millions RPS |
| RabbitMQ | Queue-based messaging | Publisher → Exchange → Queue → Consumer; simple to start; AMQP |
| NATS | Lightweight pub/sub | Subjects + subscribers; wildcard matching; at-most-once; IoT/Go-native |
