# gRPC Microservices in Go — Rules for Claude

Distilled from *gRPC Microservices in Go* by Hüseyin Babal (Manning, 2023).

---

## 1. Protobuf / .proto File Conventions

**Field naming:** Always lowercase with underscores (`user_id`, not `userId`). Protoc enforces this for multi-language codegen.

**Field numbers:**
- Reserve 1–15 for frequently-used fields (1-byte encoding); 16–2047 cost 2 bytes.
- Never reuse a field number after removal. Reserve both the number and name:
  ```proto
  reserved 4, 9;
  reserved "old_field";
  ```

**Repeated fields:** Use `repeated` for ordered multi-value fields:
```proto
repeated OrderItem items = 2;
```

**Required file header:**
```proto
syntax = "proto3";
option go_package = "github.com/org/repo/golang/order";
```

**RPC method shapes:**
```proto
// Unary
rpc Create(CreateOrderRequest) returns (CreateOrderResponse) {}

// Server streaming
rpc List(ListOrderRequest) returns (stream ListOrderResponse) {}

// Client streaming
rpc Upload(stream UploadRequest) returns (UploadResponse) {}

// Bidirectional streaming
rpc Sync(stream SyncRequest) returns (stream SyncResponse) {}
```

**oneof — mutual exclusion enforcement:**
```proto
message CreatePaymentRequest {
    oneof payment_method {
        CreditCard credit_card = 1;
        PromoCode  promo_code  = 2;
    }
}
```
- Removing a field from `oneof` is **backward-incompatible**.
- Adding a field to `oneof` is **forward-incompatible**.
- Moving a regular field into or out of `oneof` causes data loss — always breaking.

**Backward/forward compatibility rules:**
- Adding optional fields to existing messages is safe.
- Never add required fields to existing messages; implement server-side defaults instead:
  ```go
  // Upgrading server, not client: client sends old message without vat field.
  // Server defaults vat if not provided.
  vat := req.Vat
  if vat == 0 {
      vat = defaultVAT
  }
  return &CreatePaymentResponse{TotalPrice: vat + req.Price}, nil
  ```
- Upgrading client but not server: new fields are silently ignored by the old server (safe, but no-op).
- Use semantic versioning on breaking changes; bump the package path (`/v2`).

**Code generation — two output files:**

`protoc` generates two files per `.proto`:
- `order.pb.go` — message structs and serialisation
- `order_grpc.pb.go` — service interfaces, server registration, client stubs

```bash
protoc \
  -I ./proto \
  --go_out=./golang \
  --go_opt=paths=source_relative \
  --go-grpc_out=./golang \
  --go-grpc_opt=paths=source_relative \
  ./proto/order.proto
```

Flags:
- `-I` — import path where proto imports are searched
- `--go_out` — destination for message code
- `--go_opt=paths=source_relative` — preserve folder structure
- `--go-grpc_out` — destination for service/stub code

**Install the codegen plugins:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## 2. Project Layout

Use hexagonal (ports & adapters) architecture. Dependencies flow inward; outer layers never know about each other.

```
├── cmd/                                    # Entry point; dependency injection here
├── config/                                 # Env-var readers (fail fast on missing)
└── internal/
    ├── application/
    │   └── core/
    │       ├── api/                        # Application logic (orchestrates ports)
    │       └── domain/                     # Pure domain types, no framework imports
    ├── ports/                              # Go interfaces (contracts)
    └── adapters/
        ├── db/                             # GORM database adapter
        ├── grpc/                           # gRPC server adapter
        └── payment/                        # gRPC client adapter to Payment service
```

**Initialise the project:**
```bash
mkdir -p microservices/order && cd microservices/order
go mod init github.com/<username>/microservices/order
mkdir cmd config
mkdir -p internal/adapters/db
mkdir -p internal/adapters/grpc
mkdir -p internal/application/core/api
mkdir -p internal/application/core/domain
mkdir -p internal/ports
```

**Dependency injection in `cmd/main.go`:**
```go
func main() {
    dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
    if err != nil {
        log.Fatalf("Failed to connect to database. Error: %v", err)
    }
    paymentAdapter, err := payment.NewAdapter(config.GetPaymentServiceUrl())
    if err != nil {
        log.Fatalf("Failed to initialize payment stub. Error: %v", err)
    }
    application := api.NewApplication(dbAdapter, paymentAdapter)
    grpcAdapter := grpc.NewAdapter(application, config.GetApplicationPort())
    grpcAdapter.Run()
}
```

**Proto repository layout (separate repo):**
```
microservices-proto/
├── order/order.proto
├── payment/payment.proto
├── shipping/shipping.proto
└── golang/
    ├── order/          # go.mod: github.com/org/microservices-proto/golang/order
    ├── payment/        # go.mod: github.com/org/microservices-proto/golang/payment
    └── shipping/       # go.mod: github.com/org/microservices-proto/golang/shipping
```

Tag each generated sub-module independently: `golang/order/v1.2.3`. This is the Go module convention for resolving dependencies that live in repository sub-folders.

---

## 3. Configuration

Read all config from environment variables following the 12-factor app methodology. Fail loudly at startup on missing values — never silently default.

```go
// config/config.go
package config

import (
    "log"
    "os"
    "strconv"
)

func GetEnv() string {
    return getEnvironmentValue("ENV") // "development" or "production"
}

func GetDataSourceURL() string {
    return getEnvironmentValue("DATA_SOURCE_URL")
}

func GetApplicationPort() int {
    portStr := getEnvironmentValue("APPLICATION_PORT")
    port, err := strconv.Atoi(portStr)
    if err != nil {
        log.Fatalf("port: %s is invalid", portStr)
    }
    return port
}

func GetPaymentServiceUrl() string {
    return getEnvironmentValue("PAYMENT_SERVICE_URL")
}

func getEnvironmentValue(key string) string {
    if os.Getenv(key) == "" {
        log.Fatalf("%s environment variable is missing.", key)
    }
    return os.Getenv(key)
}
```

Run a service locally:
```bash
DATA_SOURCE_URL=root:verysecretpass@tcp(127.0.0.1:3306)/order \
APPLICATION_PORT=3000 \
ENV=development \
PAYMENT_SERVICE_URL=localhost:3001 \
go run cmd/main.go
```

---

## 4. Domain Model

Domain structs live in `internal/application/core/domain/`. They use JSON struct tags for serialisation and have no framework imports.

```go
// internal/application/core/domain/order.go
package domain

import "time"

type OrderItem struct {
    ProductCode string  `json:"product_code"`
    UnitPrice   float32 `json:"unit_price"`
    Quantity    int32   `json:"quantity"`
}

type Order struct {
    ID         int64       `json:"id"`
    CustomerID int64       `json:"customer_id"`
    Status     string      `json:"status"`
    OrderItems []OrderItem `json:"order_items"`
    CreatedAt  int64       `json:"created_at"`
}

func NewOrder(customerId int64, orderItems []OrderItem) Order {
    return Order{
        CreatedAt:  time.Now().Unix(),
        Status:     "Pending",
        CustomerID: customerId,
        OrderItems: orderItems,
    }
}

// TotalPrice calculates the order total from its items.
func (o *Order) TotalPrice() float32 {
    var total float32
    for _, item := range o.OrderItems {
        total += item.UnitPrice * float32(item.Quantity)
    }
    return total
}
```

---

## 5. Port Interfaces

Ports are Go interfaces in `internal/ports/`. The application depends on these interfaces, not on concrete adapters.

```go
// internal/ports/api.go
package ports

import "github.com/org/microservices/order/internal/application/core/domain"

type APIPort interface {
    PlaceOrder(order domain.Order) (domain.Order, error)
}
```

```go
// internal/ports/db.go
package ports

import "github.com/org/microservices/order/internal/application/core/domain"

type DBPort interface {
    Get(id string) (domain.Order, error)
    Save(*domain.Order) error
}
```

```go
// internal/ports/payment.go
package ports

import "github.com/org/microservices/order/internal/application/core/domain"

type PaymentPort interface {
    Charge(*domain.Order) error
}
```

---

## 6. Application Core

```go
// internal/application/core/api/api.go
package api

import (
    "github.com/org/microservices/order/internal/application/core/domain"
    "github.com/org/microservices/order/internal/ports"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Application struct {
    db      ports.DBPort
    payment ports.PaymentPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort) *Application {
    return &Application{db: db, payment: payment}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
    err := a.db.Save(&order)
    if err != nil {
        return domain.Order{}, err
    }
    paymentErr := a.payment.Charge(&order)
    if paymentErr != nil {
        // Wrap upstream error with field context before returning.
        st, _ := status.FromError(paymentErr)
        fieldErr := &errdetails.BadRequest_FieldViolation{
            Field:       "payment",
            Description: st.Message(),
        }
        badReq := &errdetails.BadRequest{}
        badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
        orderStatus := status.New(codes.InvalidArgument, "order creation failed")
        statusWithDetails, _ := orderStatus.WithDetails(badReq)
        return domain.Order{}, statusWithDetails.Err()
    }
    return order, nil
}
```

---

## 7. Database Adapter (GORM)

```go
// internal/adapters/db/db.go
package db

import (
    "fmt"
    "github.com/org/microservices/order/internal/application/core/domain"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// DB-layer structs — separate from domain structs.
// gorm.Model adds ID, CreatedAt, UpdatedAt, DeletedAt.
type Order struct {
    gorm.Model
    CustomerID int64
    Status     string
    OrderItems []OrderItem // one-to-many; GORM detects this automatically
}

type OrderItem struct {
    gorm.Model
    ProductCode string
    UnitPrice   float32
    Quantity    int32
    OrderID     uint // back-reference to Order; GORM requires this field name
}

type Adapter struct {
    db *gorm.DB
}

func NewAdapter(dataSourceUrl string) (*Adapter, error) {
    db, err := gorm.Open(mysql.Open(dataSourceUrl), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("db connection error: %v", err)
    }
    // AutoMigrate creates/updates tables at startup.
    if err := db.AutoMigrate(&Order{}, OrderItem{}); err != nil {
        return nil, fmt.Errorf("db migration error: %v", err)
    }
    return &Adapter{db: db}, nil
}

func (a Adapter) Get(id string) (domain.Order, error) {
    var orderEntity Order
    res := a.db.First(&orderEntity, id)
    var orderItems []domain.OrderItem
    for _, item := range orderEntity.OrderItems {
        orderItems = append(orderItems, domain.OrderItem{
            ProductCode: item.ProductCode,
            UnitPrice:   item.UnitPrice,
            Quantity:    item.Quantity,
        })
    }
    return domain.Order{
        ID:         int64(orderEntity.ID),
        CustomerID: orderEntity.CustomerID,
        Status:     orderEntity.Status,
        OrderItems: orderItems,
        CreatedAt:  orderEntity.CreatedAt.UnixNano(),
    }, res.Error
}

func (a Adapter) Save(order *domain.Order) error {
    var orderItems []OrderItem
    for _, item := range order.OrderItems {
        orderItems = append(orderItems, OrderItem{
            ProductCode: item.ProductCode,
            UnitPrice:   item.UnitPrice,
            Quantity:    item.Quantity,
        })
    }
    orderModel := Order{
        CustomerID: order.CustomerID,
        Status:     order.Status,
        OrderItems: orderItems,
    }
    res := a.db.Create(&orderModel)
    if res.Error == nil {
        order.ID = int64(orderModel.ID)
    }
    return res.Error
}
```

Install dependencies:
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
```

---

## 8. gRPC Server Adapter

```go
// internal/adapters/grpc/server.go
package grpc

import (
    "fmt"
    "log"
    "net"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "github.com/org/microservices-proto/golang/order"
    "github.com/org/microservices/order/config"
    "github.com/org/microservices/order/internal/ports"
)

type Adapter struct {
    api  ports.APIPort
    port int
    order.UnimplementedOrderServer // embed for forward compatibility
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
    return &Adapter{api: api, port: port}
}

func (a Adapter) Run() {
    listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
    if err != nil {
        log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
    }
    grpcServer := grpc.NewServer()
    order.RegisterOrderServer(grpcServer, a)

    // Enable reflection only in development so grpcurl can discover services.
    // Never enable in production (security + performance).
    if config.GetEnv() == "development" {
        reflection.Register(grpcServer)
    }

    if err := grpcServer.Serve(listen); err != nil {
        log.Fatalf("failed to serve grpc on port %d", a.port)
    }
}
```

```go
// internal/adapters/grpc/grpc.go
package grpc

import (
    "context"
    "github.com/org/microservices-proto/golang/order"
    "github.com/org/microservices/order/internal/application/core/domain"
)

func (a Adapter) Create(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
    var orderItems []domain.OrderItem
    for _, item := range req.OrderItems {
        orderItems = append(orderItems, domain.OrderItem{
            ProductCode: item.ProductCode,
            UnitPrice:   item.UnitPrice,
            Quantity:    item.Quantity,
        })
    }
    newOrder := domain.NewOrder(req.UserId, orderItems)
    result, err := a.api.PlaceOrder(newOrder)
    if err != nil {
        return nil, err
    }
    return &order.CreateOrderResponse{OrderId: result.ID}, nil
}
```

**Test endpoints with grpcurl:**
```bash
# Create order (requires reflection to be enabled — development only)
grpcurl \
  -d '{"user_id": 123, "order_items": [{"product_code": "prod", "quantity": 4, "unit_price": 12}]}' \
  -plaintext localhost:3000 \
  Order/Create

# Expected response:
# { "orderId": "1" }

# Test Payment service directly:
grpcurl -d '{"user_id": 123, "order_id": 12, "total_price": 32}' \
  -plaintext localhost:3001 Payment/Create
```

---

## 9. gRPC Client Adapter (Payment)

Add the generated payment stub as a dependency:
```bash
go get -u github.com/org/microservices-proto/golang/payment
# Pin to a specific version to prevent unexpected breakage:
go get -u github.com/org/microservices-proto/golang/payment@v1.0.38
```

```go
// internal/adapters/payment/payment.go
package payment

import (
    "context"
    "github.com/org/microservices-proto/golang/payment"
    "github.com/org/microservices/order/internal/application/core/domain"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type Adapter struct {
    payment payment.PaymentClient
}

func NewAdapter(paymentServiceUrl string) (*Adapter, error) {
    var opts []grpc.DialOption
    opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    conn, err := grpc.Dial(paymentServiceUrl, opts...)
    if err != nil {
        return nil, err
    }
    client := payment.NewPaymentClient(conn)
    return &Adapter{payment: client}, nil
}

func (a *Adapter) Charge(order *domain.Order) error {
    _, err := a.payment.Create(context.Background(), &payment.CreatePaymentRequest{
        UserId:     order.CustomerID,
        OrderId:    order.ID,
        TotalPrice: order.TotalPrice(),
    })
    return err
}
```

In Kubernetes, use the internal service name as the URL: `payment:8081`. Kubernetes DNS resolves it automatically — no external service registry needed.

Generated client constructor convention: `New<ServiceName>Client(conn)`.

---

## 10. Timeouts and Context Propagation

Always set a deadline before making a gRPC call. Use `context.WithTimeout` (simpler) or `context.WithDeadline` (absolute time).

```go
// WithTimeout — prefer this
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()
_, err := shippingClient.Create(ctx, req)

// WithDeadline — equivalent
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
defer cancel()
```

**Propagate the same context through the entire service chain.** The deadline travels automatically — do not create a new context at each hop.

```go
// Client sets a 3-second budget.
ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
defer cancel()
orderClient.Create(ctx, req) // 2s in Order service + 2s in Product service = 4s > 3s → timeout
```

```go
// Order service passes the SAME ctx to the downstream Product call.
func (s *server) Create(ctx context.Context, in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
    time.Sleep(2 * time.Second)
    productInfo, err := s.productClient.Get(ctx, &product.GetProductRequest{ProductId: in.ProductId})
    // ctx still carries the original 3-second deadline; no new context needed.
    ...
}
```

---

## 11. Error Handling and Status Codes

**Return a status error — never a raw error string:**
```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

return nil, status.Errorf(codes.NotFound, "order %d not found", id)
```

Returning a plain `errors.New(...)` results in `Code: Unknown` on the client, which is useless for routing or retry decisions.

**Status codes to use:**

| Code | When |
|------|------|
| `OK` | Success |
| `Cancelled` | Client cancelled (e.g., competing calls, first-wins) |
| `InvalidArgument` | Bad input — caller's fault |
| `DeadlineExceeded` | Deadline expired before operation completed |
| `NotFound` | Resource does not exist |
| `AlreadyExists` | Duplicate creation attempt |
| `PermissionDenied` | Caller not authorised for this resource |
| `ResourceExhausted` | Rate limit / quota exceeded |
| `Internal` | Server-side bug (use sparingly) |
| `Unavailable` | Transient failure; client should retry |

**Rich validation errors with field details:**
```go
import "google.golang.org/genproto/googleapis/rpc/errdetails"

fieldErr := &errdetails.BadRequest_FieldViolation{
    Field:       "user_id",
    Description: "must be greater than 0",
}
badReq := &errdetails.BadRequest{}
badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
st := status.New(codes.InvalidArgument, "invalid order request")
stWithDetails, _ := st.WithDetails(badReq)
return nil, stWithDetails.Err()
```

**Wrapping upstream errors at service boundaries** (the book's pattern — Order wraps Payment errors):
```go
paymentErr := a.payment.Charge(&order)
if paymentErr != nil {
    st, _ := status.FromError(paymentErr)
    fieldErr := &errdetails.BadRequest_FieldViolation{
        Field:       "payment",
        Description: st.Message(),
    }
    badReq := &errdetails.BadRequest{}
    badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
    orderStatus := status.New(codes.InvalidArgument, "order creation failed")
    statusWithDetails, _ := orderStatus.WithDetails(badReq)
    return domain.Order{}, statusWithDetails.Err()
}
```

**Client-side — inspect error details:**
```go
// Simple: extract code + message
st, _ := status.FromError(err)
log.Printf("code: %v, message: %s", st.Code(), st.Message())

// Rich: walk field violations
st := status.Convert(err)
var allErrors []string
for _, detail := range st.Details() {
    switch t := detail.(type) {
    case *errdetails.BadRequest:
        for _, v := range t.GetFieldViolations() {
            allErrors = append(allErrors, v.Description)
            log.Printf("field %s: %s", v.GetField(), v.GetDescription())
        }
    }
}
```

---

## 12. Resiliency: Retry, Timeout, Circuit Breaker

### Retry (transient failures only)

```go
import grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

var opts []grpc.DialOption
opts = append(opts, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
    grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted), // defaults
    grpc_retry.WithMax(5),
    grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
)))
opts = append(opts, grpc.WithInsecure())
conn, err := grpc.Dial("localhost:8080", opts...)
```

- `WithCodes` defaults are `Unavailable, ResourceExhausted`. Do not retry `InvalidArgument` — it won't fix itself.
- `WithMax`: stops as soon as a non-retryable response is received, even if max hasn't been reached.
- `WithBackoff` options: `BackoffLinear`, `BackoffExponential` (doubles each time), `BackoffLinearWithJitter`, `BackoffExponentialWithJitter` (adds randomness to reduce collision).
- There are also `WithStreamingInterceptor` variants for streaming calls.

### Circuit breaker

Use `github.com/sony/gobreaker`. Initialise once per downstream dependency.

```go
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "payment",
    MaxRequests: 3,  // allowed in half-open state
    Timeout:     4,  // seconds before open → half-open transition
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return failureRatio >= 0.6 // open circuit if 60% of requests fail
    },
    OnStateChange: func(name string, from, to gobreaker.State) {
        log.Printf("Circuit Breaker %s: %v → %v", name, from, to)
    },
})
```

**Wrap each gRPC call:**
```go
orderResponse, err := cb.Execute(func() (interface{}, error) {
    return orderClient.Create(context.Background(), req)
})
```

**As a reusable interceptor (centralises circuit-breaker logic):**
```go
// middleware/circuit_breaker.go
package middleware

import (
    "context"
    "github.com/sony/gobreaker"
    "google.golang.org/grpc"
)

func CircuitBreakerClientInterceptor(cb *gobreaker.CircuitBreaker) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{},
        cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        _, cbErr := cb.Execute(func() (interface{}, error) {
            err := invoker(ctx, method, req, reply, cc, opts...)
            if err != nil {
                return nil, err
            }
            return nil, nil
        })
        return cbErr
    }
}

// Usage:
opts = append(opts, grpc.WithUnaryInterceptor(middleware.CircuitBreakerClientInterceptor(cb)))
```

---

## 13. Load Balancing Strategies

| | Server-side LB | Client-side LB |
|---|---|---|
| **Pros** | Easy to configure; client unaware of backends | Better performance (no extra hop); no single point of failure |
| **Cons** | More latency; limited throughput can affect scalability | Hard to configure; client tracks instance health and load; language-specific |

The book uses **server-side load balancing via Kubernetes** — the Kubernetes `Service` object is the single endpoint; it proxies to pods. No extra configuration needed in the gRPC client.

---

## 14. TLS and Security

**Development only:** `insecure.NewCredentials()`. Never in production.

**Production — server TLS (mutual TLS):**
```go
func serverTLSCredentials() (credentials.TransportCredentials, error) {
    serverCert, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
    if err != nil {
        return nil, err
    }
    certPool := x509.NewCertPool()
    caCert, err := ioutil.ReadFile("cert/ca.crt")
    if err != nil {
        return nil, err
    }
    if ok := certPool.AppendCertsFromPEM(caCert); !ok {
        return nil, errors.New("failed to append CA cert")
    }
    return credentials.NewTLS(&tls.Config{
        Certificates: []tls.Certificate{serverCert},
        ClientAuth:   tls.RequireAnyClientCert, // mTLS; omit for one-way TLS
        ClientCAs:    certPool,
    }), nil
}

tlsCreds, _ := serverTLSCredentials()
grpcServer := grpc.NewServer(grpc.Creds(tlsCreds))
```

**Production — client TLS:**
```go
func clientTLSCredentials() (credentials.TransportCredentials, error) {
    clientCert, _ := tls.LoadX509KeyPair("cert/client.crt", "cert/client.key")
    certPool := x509.NewCertPool()
    caCert, _ := ioutil.ReadFile("cert/ca.crt")
    certPool.AppendCertsFromPEM(caCert)
    return credentials.NewTLS(&tls.Config{
        ServerName:   "*.microservices.dev", // must match CN used during cert generation
        Certificates: []tls.Certificate{clientCert},
        RootCAs:      certPool,
    }), nil
}
```

In Kubernetes, delegate certificate lifecycle to cert-manager rather than managing certs manually.

---

## 15. Testing

### Layers

```
Unit tests      ← fast, cheap, maximum isolation (mock all ports)
Integration     ← real database via Testcontainers, no mocks
End-to-end      ← full Docker Compose stack + real gRPC client
```

### Unit tests — mock the port interfaces

Mock by hand (for clarity):
```go
// Whatever is an interface can be mocked.
type mockedPayment struct{ mock.Mock }
func (p *mockedPayment) Charge(order *domain.Order) error {
    args := p.Called(order)
    return args.Error(0)
}

type mockedDb struct{ mock.Mock }
func (d *mockedDb) Save(order *domain.Order) error {
    args := d.Called(order)
    return args.Error(0)
}
func (d *mockedDb) Get(id string) (domain.Order, error) {
    args := d.Called(id)
    return args.Get(0).(domain.Order), args.Error(1)
}
```

Test the happy path:
```go
func Test_Should_Place_Order(t *testing.T) {
    payment := new(mockedPayment)
    db := new(mockedDb)
    payment.On("Charge", mock.Anything).Return(nil)
    db.On("Save", mock.Anything).Return(nil)

    application := NewApplication(db, payment)
    _, err := application.PlaceOrder(domain.Order{
        CustomerID: 123,
        OrderItems: []domain.OrderItem{{ProductCode: "camera", UnitPrice: 12.3, Quantity: 3}},
    })
    assert.Nil(t, err)
}
```

Test an error path:
```go
func Test_Should_Return_Error_When_Payment_Fail(t *testing.T) {
    payment := new(mockedPayment)
    db := new(mockedDb)
    payment.On("Charge", mock.Anything).Return(errors.New("insufficient balance"))
    db.On("Save", mock.Anything).Return(nil)

    application := NewApplication(db, payment)
    _, err := application.PlaceOrder(domain.Order{CustomerID: 123, OrderItems: []domain.OrderItem{{ProductCode: "bag", UnitPrice: 2.5, Quantity: 6}}})

    st, _ := status.FromError(err)
    assert.Equal(t, st.Message(), "order creation failed")
    assert.Equal(t, st.Details()[0].(*errdetails.BadRequest).FieldViolations[0].Description, "insufficient balance")
    assert.Equal(t, st.Code(), codes.InvalidArgument)
}
```

Auto-generate mocks from interfaces (instead of writing by hand):
```bash
mockery --all --keeptree
```

### Integration tests — real database via Testcontainers

```go
type OrderDatabaseTestSuite struct {
    suite.Suite
    DataSourceUrl string
}

func (o *OrderDatabaseTestSuite) SetupSuite() {
    ctx := context.Background()
    port := "3306/tcp"
    dbURL := func(port nat.Port) string {
        return fmt.Sprintf(
            "root:s3cr3t@tcp(localhost:%s)/orders?charset=utf8mb4&parseTime=True&loc=Local",
            port.Port(),
        )
    }
    req := testcontainers.ContainerRequest{
        Image:        "docker.io/mysql:8.0.30",
        ExposedPorts: []string{port},
        Env:          map[string]string{"MYSQL_ROOT_PASSWORD": "s3cr3t", "MYSQL_DATABASE": "orders"},
        WaitingFor:   wait.ForSQL(nat.Port(port), "mysql", dbURL).Timeout(30 * time.Second),
    }
    mysqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        log.Fatal("Failed to start MySQL.", err)
    }
    endpoint, _ := mysqlContainer.Endpoint(ctx, "")
    o.DataSourceUrl = fmt.Sprintf(
        "root:s3cr3t@tcp(%s)/orders?charset=utf8mb4&parseTime=True&loc=Local",
        endpoint,
    )
}

func (o *OrderDatabaseTestSuite) Test_Should_Save_Order() {
    adapter, err := NewAdapter(o.DataSourceUrl)
    o.Nil(err)
    saveErr := adapter.Save(&domain.Order{})
    o.Nil(saveErr)
}

func (o *OrderDatabaseTestSuite) Test_Should_Get_Order() {
    adapter, _ := NewAdapter(o.DataSourceUrl)
    order := domain.NewOrder(2, []domain.OrderItem{{ProductCode: "CAM", Quantity: 5, UnitPrice: 1.32}})
    adapter.Save(&order)
    ord, _ := adapter.Get(order.ID)
    o.Equal(int64(2), ord.CustomerID)
}

func TestOrderDatabaseTestSuite(t *testing.T) {
    suite.Run(t, new(OrderDatabaseTestSuite))
}
```

Notes:
- `wait.ForSQL` takes a `nat.Port`, not a string.
- `dbURL` is a **function** `func(nat.Port) string`, not a string.
- Suite assertions use the receiver (`o.Nil`, `o.Equal`), not `assert.*`.

### End-to-end tests — Docker Compose + real gRPC client

`docker-compose.yml` for the E2E stack:
```yaml
version: "3.9"
services:
  mysql:
    image: "mysql:8.0.30"
    environment:
      MYSQL_ROOT_PASSWORD: "s3cr3t"
    volumes:
      - "./init.sql:/docker-entrypoint-initdb.d/init.sql"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-ps3cr3t"]
      interval: 5s
      timeout: 5s
      retries: 20

  payment:
    depends_on:
      mysql:
        condition: service_healthy
    build: ../../payment/
    environment:
      APPLICATION_PORT: 8081
      ENV: "development"
      DATA_SOURCE_URL: "root:s3cr3t@tcp(mysql:3306)/payments?charset=utf8mb4&parseTime=True&loc=Local"

  order:
    depends_on:
      mysql:
        condition: service_healthy
    build: ../../order/
    ports:
      - "8080:8080"
    environment:
      APPLICATION_PORT: 8080
      ENV: "development"
      DATA_SOURCE_URL: "root:s3cr3t@tcp(mysql:3306)/orders?charset=utf8mb4&parseTime=True&loc=Local"
      PAYMENT_SERVICE_URL: "localhost:8081"
```

`init.sql`:
```sql
CREATE DATABASE IF NOT EXISTS payments;
CREATE DATABASE IF NOT EXISTS orders;
```

E2E test suite:
```go
type CreateOrderTestSuite struct {
    suite.Suite
    compose *testcontainers.LocalDockerCompose
}

func (c *CreateOrderTestSuite) SetupSuite() {
    composeFilePaths := []string{"resources/docker-compose.yml"}
    identifier := strings.ToLower(uuid.New().String())
    compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
    c.compose = compose
    execError := compose.WithCommand([]string{"up", "-d"}).Invoke()
    if execError.Error != nil {
        log.Fatalf("Could not run compose stack: %v", execError.Error)
    }
}

func (c *CreateOrderTestSuite) Test_Should_Create_Order() {
    var opts []grpc.DialOption
    opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    conn, err := grpc.Dial("localhost:8080", opts...)
    if err != nil {
        log.Fatalf("Failed to connect order service. Err: %v", err)
    }
    defer conn.Close()

    orderClient := order.NewOrderClient(conn)
    createResp, errCreate := orderClient.Create(context.Background(), &order.CreateOrderRequest{
        UserId: 23,
        OrderItems: []*order.OrderItem{
            {ProductCode: "CAM123", Quantity: 3, UnitPrice: 1.23},
        },
    })
    c.Nil(errCreate)

    // Also verify Get returns the same data.
    getResp, errGet := orderClient.Get(context.Background(), &order.GetOrderRequest{OrderId: createResp.OrderId})
    c.Nil(errGet)
    c.Equal(int64(23), getResp.UserId)
    c.Equal(float32(1.23), getResp.OrderItems[0].UnitPrice)
    c.Equal(int32(3), getResp.OrderItems[0].Quantity)
    c.Equal("CAM123", getResp.OrderItems[0].ProductCode)
}

func (c *CreateOrderTestSuite) TearDownSuite() {
    execError := c.compose.WithCommand([]string{"down"}).Invoke()
    if execError.Error != nil {
        log.Fatalf("Could not shutdown compose stack: %v", execError.Error)
    }
}

func TestCreateOrderTestSuite(t *testing.T) {
    suite.Run(t, new(CreateOrderTestSuite))
}
```

Run the E2E suite:
```bash
# Run from the e2e/ module directory
go test -run "^TestCreateOrderTestSuite$"
```

### Coverage

```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 16. Observability

### Distributed tracing with OpenTelemetry + Jaeger

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func newTracerProvider(url string) (*tracesdk.TracerProvider, error) {
    exp, _ := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
    return tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exp),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("order"),
        )),
    ), nil
}

// In main:
tp, _ := newTracerProvider("http://jaeger-otel.jaeger.svc.cluster.local:14278/api/traces")
otel.SetTracerProvider(tp)
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))
```

Wire into gRPC via interceptors:
```go
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

// Server
grpcServer := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))

// Client
conn, _ := grpc.Dial(addr, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
```

### Structured logging with trace context (logrus)

```go
import (
    log "github.com/sirupsen/logrus"
    "go.opentelemetry.io/otel/trace"
)

type serviceLogger struct {
    log.JSONFormatter
}

func (l serviceLogger) Format(entry *log.Entry) ([]byte, error) {
    span := trace.SpanFromContext(entry.Context)
    entry.Data["trace_id"] = span.SpanContext().TraceID().String()
    entry.Data["span_id"]  = span.SpanContext().SpanID().String()
    return l.JSONFormatter.Format(entry)
}

func init() {
    log.SetFormatter(serviceLogger{
        JSONFormatter: log.JSONFormatter{FieldMap: log.FieldMap{"msg": "message"}},
    })
    log.SetOutput(os.Stdout)
    log.SetLevel(log.InfoLevel)
}

// Always pass ctx so trace/span IDs are injected:
log.WithContext(ctx).Info("Creating order...")
```

### Metrics terminology

- **SLI**: Specific measured indicator (e.g., "request latency p95").
- **SLO**: Target for an SLI (e.g., "p95 latency < 100ms").
- **SLA**: Business commitment derived from SLOs.
- Report latency as percentiles (p50, p95, p99), not averages.

---

## 17. Kubernetes Deployment

**Multi-stage Dockerfile (produces minimal `scratch` image):**
```dockerfile
FROM golang:1.18 AS builder
WORKDIR /usr/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order ./cmd/main.go

FROM scratch
COPY --from=builder /usr/src/app/order ./order
CMD ["./order"]
```

**Deployment manifest:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: payment
  labels:
    app: payment
spec:
  replicas: 2
  selector:
    matchLabels:
      app: payment
  template:
    metadata:
      labels:
        app: payment
    spec:
      containers:
      - name: payment
        image: org/payment:v1.0.2
        env:
        - name: APPLICATION_PORT
          value: "8081"
        - name: DATA_SOURCE_URL
          value: "root:changeit@tcp(mysql:3306)/payments"
```

**Internal service (ClusterIP — intra-cluster only):**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: payment
spec:
  selector:
    app: payment
  ports:
  - port: 8081
    targetPort: 8081
  type: ClusterIP
```

**External service (LoadBalancer — cloud provider provisions load balancer):**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: order
  labels:
    app: order
spec:
  selector:
    app: order
  ports:
  - name: grpc
    port: 80
    protocol: TCP
    targetPort: 8080
  type: LoadBalancer
```

**Public ingress for gRPC (NGINX):**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/backend-protocol: GRPC  # required — without this NGINX uses HTTP/1.1
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: selfsigned-issuer
  name: order
spec:
  rules:
  - http:
      paths:
      - path: /Order          # the service name from the .proto file, not the Go package
        pathType: Prefix
        backend:
          service:
            name: order
            port:
              number: 8080
  tls:
  - hosts:
    - ingress.local
    secretName: order-tls
```

**Install NGINX ingress controller:**
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install nginx-ingress ingress-nginx/ingress-nginx
```

**Install cert-manager:**
```bash
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --version v1.10.0 --set installCRDs=true
```

**Self-signed ClusterIssuer:**
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
```

---

## 18. CI/CD and Proto Automation

Maintain `.proto` files in a dedicated repository (`microservices-proto`). Trigger code generation on version tags, not every push.

```yaml
# .github/workflows/proto.yml
name: "Protocol Buffer Go Stubs Generation"
on:
  push:
    tags:
      - "v*"
jobs:
  protoc:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: ["order", "payment", "shipping"]
    steps:
      - uses: actions/setup-go@v2
        with:
          go-version: 1.17
      - uses: actions/checkout@v2
      - name: Extract Release Version
        run: echo "RELEASE_VERSION=${GITHUB_REF#refs/*/}" >> $GITHUB_ENV
      - name: Generate for Golang
        run: |
          chmod +x "${GITHUB_WORKSPACE}/run.sh"
          ./run.sh ${{ matrix.service }} ${{ env.RELEASE_VERSION }}
```

`run.sh`:
```bash
#!/bin/bash
SERVICE_NAME=$1
RELEASE_VERSION=$2

sudo apt-get install -y protobuf-compiler golang-goprotobuf-dev
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc \
  --go_out=./golang --go_opt=paths=source_relative \
  --go-grpc_out=./golang --go-grpc_opt=paths=source_relative \
  ./${SERVICE_NAME}/*.proto

cd golang/${SERVICE_NAME}
go mod init github.com/org/microservices-proto/golang/${SERVICE_NAME} || true
go mod tidy
cd ../../

git config --global user.email "ci@example.com"
git config --global user.name "CI Bot"
git add . && git commit -am "proto update" || true
git tag -fa golang/${SERVICE_NAME}/${RELEASE_VERSION} -m "golang/${SERVICE_NAME}/${RELEASE_VERSION}"
git push origin refs/tags/golang/${SERVICE_NAME}/${RELEASE_VERSION}
```

Consuming a specific version:
```bash
go get -u github.com/org/microservices-proto/golang/order@v1.2.3
```

---

## 19. Anti-Patterns and Pitfalls

| Anti-pattern | What to do instead |
|---|---|
| Shared request/response models across services | Generate per-service from proto; a little duplication is cheaper than wrong abstraction |
| Returning `errors.New(...)` from a gRPC handler | Always `status.Errorf(codes.X, ...)` — raw errors produce `Code: Unknown` |
| Retry on all error codes | Retry only `Unavailable` and `ResourceExhausted`; `InvalidArgument` won't fix itself |
| gRPC calls without a timeout | Always `context.WithTimeout`; `context.WithDeadline` for absolute times |
| Creating a new context at each service hop | Propagate the original context — deadline travels automatically |
| Reusing removed proto field numbers | Reserve with `reserved` keyword (number and name) |
| Moving a field into/out of `oneof` | Always breaking — version the API instead |
| Silently ignoring empty env vars | `log.Fatalf("%s environment variable is missing.", key)` at startup |
| Enabling gRPC reflection in production | `if config.GetEnv() == "development" { reflection.Register(grpcServer) }` |
| Business logic in the gRPC adapter | gRPC adapter converts types and delegates to `APIPort`; logic lives in `application/core/api` |
| No circuit breaker on downstream calls | Wrap with `gobreaker`; open at 60% failure ratio to prevent cascading failures |
| Deploying without observability | Wire `otelgrpc` interceptors and logrus trace formatter from day one |
| Using `grpc.WithInsecure()` in production | Use `grpc.WithTransportCredentials(tlsCreds)` with a proper cert chain |

---

## Quick Reference

```go
// Full server startup with OTel
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
)
order.RegisterOrderServer(grpcServer, adapter)
if config.GetEnv() == "development" {
    reflection.Register(grpcServer)
}

// Full client dial (production, with retry + OTel)
creds, _ := clientTLSCredentials()
conn, _ := grpc.Dial(addr,
    grpc.WithTransportCredentials(creds),
    grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
        grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
        grpc_retry.WithMax(5),
        grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
    )),
    grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
)

// Error return
return nil, status.Errorf(codes.NotFound, "order %d not found", orderID)

// Context with timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```
