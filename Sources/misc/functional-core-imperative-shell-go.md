# Mastering Functional Core/Imperative Shell in Go: A Pragmatic Guide to Clean Architecture
[From Farid Ghoorchian Sep 2, 2025](https://medium.com/@farid-ghr/mastering-functional-core-imperative-shell-in-go-a-pragmatic-guide-to-clean-architecture-dacab227497f)


### Why Functional Core/Imperative Shell Matters

Software engineering trends come and go, but one thing never changes: we all want **clean, testable, and maintainable code**.

That's where **Functional Core/Imperative Shell (FC/IS)** comes in — a paradigm that mixes the rigor of functional programming with the practicality of Go's imperative style. Popularized by Gary Bernhardt's legendary _Boundaries_ talk, FC/IS pushes all the messy side effects (I/O, database calls, logging) to the **shell**, while keeping the **core** pure, deterministic, and trivial to test.

This post is a distilled version of a deep-dive conversation I had with a fellow Go developer on applying FC/IS in real projects. We'll cover:

*   The basics of FC/IS in Go
*   Common pitfalls like tight coupling
*   How to introduce an orchestrator layer
*   FC/IS vs. OOP Clean Architecture
*   Handling multiple side effects without creating "god objects"
*   A sample project structure

And yes — we'll make it concrete with a **shopping cart example**.

### FC/IS in Go: The Basics

Think of FC/IS as two concentric layers:

*   **Functional Core** → pure business logic. No I/O, no globals, no hidden state.
*   **Imperative Shell** → the messy real world: reading JSON, handling HTTP requests, writing to databases.

Go is a natural fit: functions return values and errors cleanly, and immutability can be approximated with value passing.

### Example: Calculating a Shopping Cart Total

**Core (pure functions):**

Copy`package core  type Item struct {     Name     string     Price    float64     Quantity int     Discount float64 }  type Cart struct {     Items []Item }  func ValidateItem(item Item) (Item, error) {     if item.Price < 0 || item.Quantity < 0 || item.Discount < 0 || item.Discount > 100 {         return Item{}, fmt.Errorf("invalid item: %v", item)     }     return item, nil }  func ApplyDiscount(item Item) float64 {     return item.Price * (1 - item.Discount/100) * float64(item.Quantity) }  func CalculateTotal(cart Cart) (float64, error) {     var total float64     for _, item := range cart.Items {         validated, err := ValidateItem(item)         if err != nil {             return 0, err         }         total += ApplyDiscount(validated)     }     return total, nil }`

Shell (simple `main`):

Copy`package main  import (     "encoding/json"     "fmt"     "os" )  func main() {     data, err := os.ReadFile("cart.json")     if err != nil {         fmt.Printf("Error: %v\n", err)         os.Exit(1)     }      var cart core.Cart     json.Unmarshal(data, &cart)      total, err := core.CalculateTotal(cart)     if err != nil {         fmt.Printf("Error: %v\n", err)         os.Exit(1)     }      fmt.Printf("Total: $%.2f\n", total) }`

Here, the **core stays pure**, while the **shell handles side effects**. Nice and simple. But as projects grow, we need more structure.

### The First Pitfall: Tight Coupling in Handlers

In a web app, if you put shell logic directly into HTTP handlers, you quickly get:

*   **Duplication** (same logic in CLI, gRPC, etc.)
*   **Poor testability** (handlers hard to unit test)
*   **Rigidity** (hard to swap entry points)

### Enter the Orchestrator

The fix is to add an **orchestrator layer** — a reusable shell that coordinates workflows, while adapters (HTTP, CLI, etc.) stay thin.

Copy`package services  import "context"  type CartRequest struct {     Items []core.Item }  type CartResponse struct {     Total float64     Error error }  func ProcessCart(ctx context.Context, req CartRequest) CartResponse {     cart := core.Cart{Items: req.Items}     total, err := core.CalculateTotal(cart)     if err != nil {         return CartResponse{Error: err}     }     return CartResponse{Total: total} }`

Now both HTTP and CLI can just call `ProcessCart`. No duplication.

### Wait… Isn't This Just Clean Architecture?

Kind of. As we add orchestrators and inject dependencies (e.g., loggers, repos), the structure starts looking suspiciously like Robert C. Martin's Clean Architecture.

**Similarities:**

*   Clear separation of concerns
*   Dependency inversion via interfaces

**Differences:**

*   FC/IS is function-first, not object-first
*   The "core" is stateless and pure
*   Go idioms (errors as values, not exceptions) make it feel lighter
*   It's simpler: just two layers instead of concentric rings

To keep it _Go-ish_, avoid turning everything into structs with methods. Functions are usually enough.

### Avoiding the "God Orchestrator"

One common anti-pattern: shoving **all the side effects** into a single orchestrator (user auth, inventory checks, payment, etc.).

That way lies madness: bloated services, fragile tests, painful maintenance.

### The Fix: Focused Services + Composition

Instead, break things down:

*   `UserService` → authentication
*   `InventoryService` → stock validation
*   `CartService` → cart totals
*   `CheckoutOrchestrator` → sequences them

Copy`type CheckoutOrchestrator struct {     userSvc      *UserService     inventorySvc *InventoryService     cartSvc      *CartService }  func (o *CheckoutOrchestrator) Execute(ctx context.Context, userID string, cart core.Cart) (float64, error) {     if _, err := o.userSvc.Authenticate(ctx, userID); err != nil {         return 0, err     }     if err := o.inventorySvc.ValidateInventory(ctx, cart.Items); err != nil {         return 0, err     }     return o.cartSvc.ProcessCart(ctx, cart) }`

This keeps responsibilities clear, side effects minimal, and tests focused.

### A Practical Project Layout

Here's how you might organize a real Go app with FC/IS:

Copy`cartapp/ ├── go.mod ├── cmd/ │   ├── http/main.go  # HTTP entry point │   └── cli/main.go   # CLI entry point ├── internal/ │   ├── core/         # Pure functions │   │   └── cart.go │   ├── services/     # Orchestrators │   │   ├── cart_service.go │   │   ├── user_service.go │   │   ├── inventory_service.go │   │   └── checkout_orchestrator.go │   ├── repositories/ # DB interfaces + impls │   │   ├── user_repo.go │   │   └── merch_repo.go │   ├── adapters/     # Interface-specific │   │   ├── http/handler.go │   │   └── cli/cmd.go │   └── logging/logger.go └── tests/            # E2E tests`

The rule of thumb: **core stays pure, services orchestrate, adapters talk to the outside world.**

### Final Thoughts

FC/IS isn't theory — it's a practical way to structure Go applications. By keeping your **core pure** and your **shell imperative**, you gain:

*   **Testability** (pure functions are trivial to test)
*   **Maintainability** (side effects isolated at the edges)
*   **Flexibility** (reuse logic across HTTP, CLI, gRPC, etc.)

The key is balance: don't let orchestrators balloon into monsters, and don't over-engineer for small projects.

If you've never tried FC/IS, start small: refactor one service into a pure core with an orchestrator shell. You'll feel the difference.

What's your take: Is FC/IS simpler than Clean Architecture in Go, or just a different flavor of it? Share your thoughts in the comments.

_Thanks for reading! If this sparked ideas, follow me on Medium for more deep dives into Go architecture._

_Claps and shares are always appreciated 🚀_
