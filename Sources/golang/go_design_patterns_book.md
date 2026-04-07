# Go Design Patterns

*Mario Castro Contreras — Packt Publishing*

---

## Preface

Go is a statically typed, compiled programming language with a focus on simplicity and performance. It was designed at Google to address weaknesses in other languages while retaining their positive characteristics.

This book covers the most important design patterns known in the industry and adapts them to the Go programming language. We will start with the basic principles of Go and the most well-known creational patterns (Singleton, Builder, Factory, Abstract Factory, Prototype). Then, we will move to structural patterns (Composite, Adapter, Bridge, Proxy, Facade, Decorator, Flyweight) and behavioral patterns (Strategy, Chain of Responsibility, Command, Template, Memento, Interpreter, Visitor, State, Mediator, Observer). Finally, we will cover concurrency patterns specific to Go (Barrier, Future, Pipeline, Workers Pool, and a concurrent Publish/Subscribe implementation).

The approach throughout the book is Test-Driven Development (TDD): we write acceptance criteria and failing tests before writing the implementation.

---

## Chapter 1. Ready... Steady... Go!

### A little bit of history

Go was born at Google in 2007 and was publicly announced in 2009. It was created by Robert Griesemer, Rob Pike, and Ken Thompson. The main design goals were:

- A statically typed language with expressive power approaching dynamic languages
- Compiled but with fast compilation times
- Networked and multicore computing support built-in
- Garbage collected

Go 1.0 was released in 2012, and the language has maintained backward compatibility with that version ever since.

### Installing Go

Go is available for Linux, macOS, and Windows. Download the official package from https://golang.org/dl/ and follow the installation instructions for your platform.

After installation, verify it works:

```bash
go version
```

The output should look like:

```
go version go1.x.x linux/amd64
```

### The GOPATH

The `GOPATH` is a workspace directory that contains your Go source files, compiled binaries, and downloaded packages. It defaults to `$HOME/go` on Unix systems.

The workspace has three subdirectories:

- `src/` — Go source files
- `pkg/` — compiled package objects
- `bin/` — compiled executables

### Types

Go has the following built-in types:

- Boolean: `bool` (`true` or `false`)
- Numeric: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`
- Floating-point: `float32`, `float64`
- Complex: `complex64`, `complex128`
- String: `string`
- Byte: `byte` (alias for `uint8`)
- Rune: `rune` (alias for `int32`)

### Variables

Variables can be declared in several ways:

```go
var x int          // zero value: 0
var x int = 5      // explicit type with value
x := 5             // short declaration (infers type)
```

The zero values are:
- `0` for numeric types
- `false` for bool
- `""` for string
- `nil` for pointers, functions, interfaces, slices, channels, and maps

### Constants

```go
const Pi = 3.14159
const (
    StatusOK = 200
    StatusNotFound = 404
)
```

### Functions

```go
func add(a, b int) int {
    return a + b
}

// Multiple return values
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

Named return values and `defer`:

```go
func readFile(name string) (content string, err error) {
    f, err := os.Open(name)
    if err != nil {
        return
    }
    defer f.Close()
    // ...
    return
}
```

### Arrays, slices, and maps

Arrays have a fixed size:

```go
var arr [3]int
arr[0] = 1
arr[1] = 2
arr[2] = 3
```

Slices are dynamic:

```go
s := []int{1, 2, 3}
s = append(s, 4)
```

Maps are key-value stores:

```go
m := map[string]int{
    "one": 1,
    "two": 2,
}
m["three"] = 3
delete(m, "one")
```

### Interfaces

An interface defines a set of methods. Any type that implements those methods satisfies the interface — there is no explicit declaration of intent.

```go
type Animal interface {
    Speak() string
}

type Dog struct{}

func (d Dog) Speak() string {
    return "Woof!"
}

type Cat struct{}

func (c Cat) Speak() string {
    return "Meow!"
}
```

### Structs

Structs are composite types:

```go
type Person struct {
    Name string
    Age  int
}

p := Person{Name: "Alice", Age: 30}
fmt.Println(p.Name)
```

Structs can embed other structs:

```go
type Employee struct {
    Person
    Company string
}

e := Employee{
    Person:  Person{Name: "Bob", Age: 25},
    Company: "Acme",
}
fmt.Println(e.Name) // promoted field
```

### Pointers

```go
x := 5
p := &x    // p is a pointer to x
*p = 10    // dereference to set the value
fmt.Println(x) // 10
```

### Testing in Go

Go has a built-in testing package. Test files end in `_test.go` and test functions begin with `Test`:

```go
package mypackage

import "testing"

func TestAdd(t *testing.T) {
    result := add(2, 3)
    if result != 5 {
        t.Errorf("Expected 5, got %d", result)
    }
}
```

Run tests with:

```bash
go test ./...
go test -v ./...
```

### Test-Driven Development (TDD)

TDD follows three steps:

1. Write a failing test that describes the behavior you want.
2. Write the minimum code to make the test pass.
3. Refactor the code while keeping tests green.

This book uses TDD throughout to demonstrate how patterns emerge naturally from requirements.

### JSON

Go has excellent JSON support in the standard library:

```go
import "encoding/json"

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// Marshal (Go → JSON)
p := Person{Name: "Alice", Age: 30}
data, err := json.Marshal(p)

// Unmarshal (JSON → Go)
var p2 Person
err = json.Unmarshal(data, &p2)
```

### Go tools

Useful built-in tools:

```bash
go build ./...     # compile
go test ./...      # run tests
go fmt ./...       # format code
go vet ./...       # static analysis
go doc fmt         # view documentation
go get <pkg>       # download package
```

---

## Chapter 2. Creational Patterns

Creational patterns deal with object creation. They abstract the instantiation process to make a system independent of how its objects are created.

### Singleton design pattern

#### Description

The Singleton pattern ensures that a type has only one instance and provides a global point of access to it.

#### Objectives

- Have a single instance of a particular type
- Provide a global access point to that single instance

#### A unique counter

We want a counter that can only have one instance in the application. Any part of the code accessing the counter will use the same value.

#### Acceptance criteria

1. When no instance has been created, a new one is created and a zero value counter is returned.
2. If an instance has already been created, that instance is returned and the counter holds any previously stored values.
3. If the method `AddOne` is called, the counter must be incremented by 1.

#### Unit tests

```go
var singleton *singleton

func TestGetInstance(t *testing.T) {
    counter1 := GetInstance()
    if counter1 == nil {
        t.Error("Expected pointer to Singleton after calling GetInstance(), not nil")
    }

    expectedCounter := counter1
    counter1.AddOne()

    counter2 := GetInstance()
    if counter2 != expectedCounter {
        t.Error("Expected same instance in counter2 but got a different one")
    }

    if counter2.Val != 1 {
        t.Errorf("Counter in counter2 should be 1 but was %d\n", counter2.Val)
    }

    counter2.AddOne()
    counter3 := GetInstance()

    if counter3.Val != 2 {
        t.Errorf("Counter in counter3 should be 2 but was %d\n", counter3.Val)
    }
}
```

#### Implementation

```go
package creational

type singleton struct {
    count int
}

var instance *singleton

func GetInstance() *singleton {
    if instance == nil {
        instance = new(singleton)
    }
    return instance
}

func (s *singleton) AddOne() {
    s.count++
}
```

> **Note:** This basic implementation is not concurrent safe. See Chapter 8 for a concurrent-safe Singleton.

### Builder design pattern

#### Description

The Builder pattern constructs complex objects step by step. It separates the construction of an object from its representation.

#### Objectives

- Abstract complex creations so that object creation is separated from the object user
- Create an object step by step by filling its fields and creating the embedded objects
- Reuse the object creation algorithm between many objects

#### A vehicle manufacturer

We want to manufacture cars and motorcycles. Both share some assembly steps but differ in others.

#### Acceptance criteria

1. We must have a manufacturing type that constructs everything a vehicle needs.
2. When using a car builder, a `VehicleProduct` with four wheels, five seats, and a car structure is returned.
3. When using a motorcycle builder, a `VehicleProduct` with two wheels, two seats, and a motorcycle structure is returned.

#### Implementation

```go
package creational

const (
    LuxuryCarType = 1
    FamilyCarType = 2
)

type BuildProcess interface {
    SetWheels() BuildProcess
    SetSeats() BuildProcess
    SetStructure() BuildProcess
    GetVehicle() VehicleProduct
}

type ManufacturingDirector struct{}

func (f *ManufacturingDirector) Construct(b BuildProcess) {
    b.SetSeats().SetStructure().SetWheels()
}

type VehicleProduct struct {
    Wheels    int
    Seats     int
    Structure string
}

type CarBuilder struct {
    v VehicleProduct
}

func (c *CarBuilder) SetWheels() BuildProcess {
    c.v.Wheels = 4
    return c
}

func (c *CarBuilder) SetSeats() BuildProcess {
    c.v.Seats = 5
    return c
}

func (c *CarBuilder) SetStructure() BuildProcess {
    c.v.Structure = "Car"
    return c
}

func (c *CarBuilder) GetVehicle() VehicleProduct {
    return c.v
}

type BikeBuilder struct {
    v VehicleProduct
}

func (b *BikeBuilder) SetWheels() BuildProcess {
    b.v.Wheels = 2
    return b
}

func (b *BikeBuilder) SetSeats() BuildProcess {
    b.v.Seats = 2
    return b
}

func (b *BikeBuilder) SetStructure() BuildProcess {
    b.v.Structure = "Motorcycle"
    return b
}

func (b *BikeBuilder) GetVehicle() VehicleProduct {
    return b.v
}
```

### Factory design pattern

#### Description

The Factory pattern provides an interface for creating objects without specifying the exact type of the object that will be created. It delegates object creation to subclasses or factory functions.

#### Objectives

- Make the creation of new instances of types flexible
- Group families of objects to get a family object creator
- Abstract any user from the knowledge of the concrete implementations they use

#### A payment method factory

We want to create different payment methods: cash and credit card.

#### Acceptance criteria

1. Create each payment method using a factory function called `GetPaymentMethod(paymentType int)`.
2. When a credit card payment method is created, it must return the message `Paid X using credit card`.
3. When a cash payment method is created, it must return the message `Paid X using cash`.
4. If a payment method is requested that is not recognized, an error must be returned.

#### Implementation

```go
package creational

import (
    "errors"
    "fmt"
)

const (
    Cash      = 1
    DebitCard = 2
)

type PaymentMethod interface {
    Pay(amount float32) string
}

func GetPaymentMethod(m int) (PaymentMethod, error) {
    switch m {
    case Cash:
        return new(CashPM), nil
    case DebitCard:
        return new(DebitCardPM), nil
    default:
        return nil, errors.New(fmt.Sprintf("Payment method %d not recognized\n", m))
    }
}

type CashPM struct{}
type DebitCardPM struct{}

func (c *CashPM) Pay(amount float32) string {
    return fmt.Sprintf("%0.2f paid using cash\n", amount)
}

func (c *DebitCardPM) Pay(amount float32) string {
    return fmt.Sprintf("%#.2f paid using debit card\n", amount)
}
```

### Abstract Factory design pattern

#### Description

The Abstract Factory provides an interface for creating families of related objects without specifying their concrete classes. It is like a factory of factories.

#### Objectives

- Provide a new layer of encapsulation for factory methods that return a common interface for all factories
- Group families of object creators so that only one Abstract Factory is used

#### A vehicle factory store

We want to create vehicle factories: one for motorbikes and one for cars. Each factory produces a vehicle with a motor and the body chassis.

#### Implementation

```go
package creational

import "fmt"

type Vehicle interface {
    GetWheels() int
    GetSeats() int
}

type VehicleFactory interface {
    NewVehicle(v int) (Vehicle, error)
}

const (
    CarFactoryType       = 1
    MotorbikeFactoryType = 2
    Luxury               = 1
    Family               = 2
    Motocross            = 1
    Cruise               = 2
)

func GetVehicleFactory(f int) (VehicleFactory, error) {
    switch f {
    case CarFactoryType:
        return new(carFactory), nil
    case MotorbikeFactoryType:
        return new(motorbikeFactory), nil
    }
    return nil, fmt.Errorf("Factory with id %d not recognized\n", f)
}

type carFactory struct{}

func (c *carFactory) NewVehicle(v int) (Vehicle, error) {
    switch v {
    case Luxury:
        return &luxuryCar{}, nil
    case Family:
        return &familyCar{}, nil
    }
    return nil, fmt.Errorf("Vehicle with id %d not recognized\n", v)
}

type motorbikeFactory struct{}

func (m *motorbikeFactory) NewVehicle(v int) (Vehicle, error) {
    switch v {
    case Motocross:
        return &motocrossMotorbike{}, nil
    case Cruise:
        return &cruiseMotorbike{}, nil
    }
    return nil, fmt.Errorf("Vehicle with id %d not recognized\n", v)
}
```

### Prototype design pattern

#### Description

The Prototype pattern creates new objects by copying an existing object (the prototype). The copy is called a clone.

#### Objectives

- Maintain a collection of objects that will be cloned to create new instances
- Provide a default value of some type to be used as a start point for other instances

#### A shirts shop

We want a store that provides ready-to-print shirts. Each shirt prototype has a color, price, and SKU.

#### Acceptance criteria

1. Have a shirt cloner object with a `GetClone(s int)` method.
2. When asking for the white shirt item (id 1), a clone of the white shirt prototype is returned.
3. When asking for the black shirt item (id 2), a clone of the black shirt prototype is returned.

#### Implementation

```go
package creational

import "fmt"

const (
    White = 1
    Black = 2
    Blue  = 3
)

type ShirtCloner interface {
    GetClone(s int) (ItemInfoGetter, error)
}

func GetShirtCloner() ShirtCloner {
    return &ShirtsCache{}
}

type ShirtsCache struct{}

func (s *ShirtsCache) GetClone(s int) (ItemInfoGetter, error) {
    switch s {
    case White:
        newItem := *whitePrototype
        return &newItem, nil
    case Black:
        newItem := *blackPrototype
        return &newItem, nil
    case Blue:
        newItem := *bluePrototype
        return &newItem, nil
    default:
        return nil, fmt.Errorf("Shirt model %d not recognized\n", s)
    }
}

type ItemInfoGetter interface {
    GetInfo() string
}

type ShirtColor byte

type Shirt struct {
    Price float32
    SKU   string
    Color ShirtColor
}

func (s *Shirt) GetInfo() string {
    return fmt.Sprintf("Shirt with SKU '%s' and Color id %d that costs %f\n",
        s.SKU, s.Color, s.Price)
}

var whitePrototype *Shirt = &Shirt{
    Price: 15.00,
    SKU:   "empty",
    Color: White,
}

var blackPrototype *Shirt = &Shirt{
    Price: 16.00,
    SKU:   "empty",
    Color: Black,
}

var bluePrototype *Shirt = &Shirt{
    Price: 17.00,
    SKU:   "empty",
    Color: Blue,
}
```

---

## Chapter 3. Structural Patterns — Composite, Adapter, and Bridge

Structural patterns explain how to assemble objects and classes into larger structures, keeping these structures flexible and efficient.

### Composite design pattern

#### Description

The Composite pattern composes objects into tree structures to represent part-whole hierarchies. It lets clients treat individual objects and compositions of objects uniformly.

#### Objectives

- Learn how to create tree-like structures of objects
- Design a structure built from small objects with different parts

#### A soldier

We want to model a swimmer who is also a shooter. Instead of creating a single large interface, we compose smaller interfaces.

#### Implementation

```go
package structural

type Athlete struct{}

func (a *Athlete) Train() {
    println("Training")
}

type CompositeSwimmerA struct {
    MyAthlete Athlete
    MySwim    func()
}

type Swimmer interface {
    Swim()
}

type Trainer interface {
    Train()
}

type SwimmerImplementor struct{}

func (s *SwimmerImplementor) Swim() {
    println("Swimming!")
}

type CompositeSwimmerB struct {
    Trainer
    Swimmer
}
```

#### A tree of products

```go
package structural

type SkuFinder interface {
    FindByReferenceName(name string) Article
}

type Article struct {
    Name  string
    Price float32
}

type Products struct {
    All []Article
}

func (p *Products) FindByReferenceName(name string) Article {
    for _, product := range p.All {
        if product.Name == name {
            return product
        }
    }
    return Article{}
}
```

### Adapter design pattern

#### Description

The Adapter pattern allows incompatible interfaces to work together. It converts the interface of a class into another interface that clients expect.

#### Objectives

- Translate one interface into the one an existing code expects
- Provide compatibility between parts that were built differently

#### An incompatible interface with an adapter

We have a `LegacyPrinter` that prints with a specific signature. We want to adapt it to a `ModernPrinter` interface.

#### Acceptance criteria

1. Create an `Adapter` type that implements `ModernPrinter` and uses `LegacyPrinter` internally.
2. The new `PrintStored()` method on the `Adapter` type must delegate printing to `LegacyPrinter`.

#### Implementation

```go
package structural

import "fmt"

type LegacyPrinter interface {
    Print(s string) string
}

type MyLegacyPrinter struct{}

func (m *MyLegacyPrinter) Print(s string) (newMsg string) {
    newMsg = fmt.Sprintf("Legacy Printer: %s\n", s)
    println(newMsg)
    return
}

type ModernPrinter interface {
    PrintStored() string
}

type PrinterAdapter struct {
    OldPrinter LegacyPrinter
    Msg        string
}

func (p *PrinterAdapter) PrintStored() (newMsg string) {
    if p.OldPrinter != nil {
        withMsg := fmt.Sprintf("Adapter: %s", p.Msg)
        newMsg = p.OldPrinter.Print(withMsg)
    } else {
        newMsg = p.Msg
    }
    return
}
```

### Bridge design pattern

#### Description

The Bridge pattern decouples an abstraction from its implementation so that the two can vary independently.

#### Objectives

- Decouple abstraction (what it does) from implementation (how it does it)
- Provide the flexibility to change implementations without changing the client code

#### Two printers and two printing mechanisms

We have two printers (normal and Gutenberg) and want to bridge them with the printing mechanism.

#### Implementation

```go
package structural

import "fmt"

type PrintAPI interface {
    PrintMessage(msg string)
}

type PrintImpl1 struct{}

func (p *PrintImpl1) PrintMessage(msg string) {
    fmt.Printf("%s\n", msg)
}

type PrintImpl2 struct{}

func (p *PrintImpl2) PrintMessage(msg string) {
    fmt.Printf(":%s:\n", msg)
}

type NormalPrinter struct {
    Msg     string
    Printer PrintAPI
}

func (c *NormalPrinter) Print() error {
    c.Printer.PrintMessage(c.Msg)
    return nil
}

type PacktPrinter struct {
    Msg     string
    Printer PrintAPI
}

func (c *PacktPrinter) Print() error {
    c.Printer.PrintMessage(fmt.Sprintf("Packt: %s", c.Msg))
    return nil
}
```

---

## Chapter 4. Structural Patterns — Proxy, Facade, Decorator, and Flyweight

### Proxy design pattern

#### Description

The Proxy pattern provides a surrogate or placeholder for another object to control access to it.

#### Objectives

- Control access to a type
- Provide a wrapper over a type to hide complexity from the user
- Provide an extra layer of security to the original type

#### An authorization proxy

We want an authorization middleware that wraps user access and checks permissions before allowing operations.

#### Acceptance criteria

1. Create a proxy type that receives a user and checks if they have the right privileges.
2. If the user has admin privileges, they can perform all operations.
3. If the user does not have admin privileges, the access is denied.

#### Implementation

```go
package structural

import "fmt"

type User struct {
    ID int32
}

type AccessListChecker interface {
    MayAccess(user *User) bool
}

type Proxy struct {
    SomeDatabase SomeDatabase
}

type SomeDatabase struct{}

func (s *SomeDatabase) GetData() string {
    return "some data"
}

func (p *Proxy) GetData(user *User) string {
    if user.ID == 1 {
        return p.SomeDatabase.GetData()
    }
    return fmt.Sprintf("User %d does not have access", user.ID)
}
```

### Facade design pattern

#### Description

The Facade pattern provides a simplified interface to a complex subsystem. It hides the complexity and provides a simpler API.

#### Objectives

- Use a Facade when you want to provide a simple interface to a complex body of code
- Allow many implementations and algorithms behind a wall
- Group a set of poorly designed APIs into a single better-designed API

#### A OpenSSL alternative

We will make a simplified interface to generate an accepted and secure OpenSSL-like API.

#### Implementation

```go
package structural

import (
    "crypto/sha256"
    "fmt"
)

type SecurityFacade struct{}

func (s *SecurityFacade) GenerateSecurePassword(password string) string {
    hash := sha256.New()
    hash.Write([]byte(password))
    return fmt.Sprintf("%x", hash.Sum(nil))
}
```

### Decorator design pattern

#### Description

The Decorator pattern attaches additional responsibilities to an object dynamically. It provides a flexible alternative to subclassing for extending functionality.

#### Objectives

- Extend the functionality of some types without the risk of breaking the parent types
- Add functionality to some code at runtime
- Provide a good alternative to subclassing

#### A server middleware with logging and authentication

We want to add logging and authentication as decorators to an HTTP server handler.

#### Acceptance criteria

1. Have a simple HTTP server with one handler that returns `Hello, World!`.
2. Create a logging decorator that prints the method and the request URL.
3. Create a basic authentication decorator that checks for a valid user.

#### Implementation

```go
package structural

import (
    "fmt"
    "log"
    "net/http"
)

type MyServer struct{}

func (m *MyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, World!")
}

type Logger struct {
    Inner http.Handler
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("Before serving request: %s %s\n", r.Method, r.RequestURI)
    l.Inner.ServeHTTP(w, r)
    log.Printf("After serving request: %s %s\n", r.Method, r.RequestURI)
}

type Authentication struct {
    Inner    http.Handler
    Username string
    Password string
}

func (a *Authentication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    u, p, ok := r.BasicAuth()
    if !ok || u != a.Username || p != a.Password {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    a.Inner.ServeHTTP(w, r)
}
```

### Flyweight design pattern

#### Description

The Flyweight pattern reduces the memory cost of creating and manipulating a large number of similar objects by sharing as much data as possible.

#### Objectives

- Cache data for reuse between many small objects
- Have a pool of objects to use for a wide variety of requests

#### A team of knights with a pool of memory

We want a pool of swords that can be reused among many knights.

#### Implementation

```go
package structural

import "fmt"

type Team struct {
    ID      uint64
    Name    string
    Shield  Shield
    Players []Player
    History []HistoricalData
}

type Shield struct {
    ID    uint64
    Name  string
    Color string
    History []HistoricalData
}

type Player struct {
    ID       uint64
    Name     string
    Position string
}

type HistoricalData struct {
    Year          uint32
    LeagueResults []match
}

type match struct {
    Date          string
    VisitorID     uint64
    LocalID       uint64
    LocalScore    byte
    VisitorScore  byte
    LocalShoots   uint16
    VisitorShoots uint16
}

type teamFlyweightFactory struct {
    createdTeams map[int]*Team
}

func (f *teamFlyweightFactory) GetTeam(id int) *Team {
    if f.createdTeams[id] != nil {
        return f.createdTeams[id]
    }
    newTeam := teams[id]
    f.createdTeams[id] = &newTeam
    return f.createdTeams[id]
}

func (f *teamFlyweightFactory) ExtractedTeams() int {
    return len(f.createdTeams)
}

func NewTeamFactory() teamFlyweightFactory {
    return teamFlyweightFactory{
        createdTeams: make(map[int]*Team),
    }
}

var createdTeams int

var teams = [...]Team{
    {
        ID:   1,
        Name: fmt.Sprintf("Team A"),
    },
    {
        ID:   2,
        Name: fmt.Sprintf("Team B"),
    },
}
```

---

## Chapter 5. Behavioral Patterns — Strategy, Chain of Responsibility, and Command

Behavioral patterns characterize the ways in which classes or objects interact and distribute responsibility.

### Strategy design pattern

#### Description

The Strategy pattern defines a family of algorithms, encapsulates each one, and makes them interchangeable. It lets the algorithm vary independently from clients that use it.

#### Objectives

- Provide several algorithms to achieve some specific functionality
- All types that implement the Strategy interface must be capable of being used in the same place, interchangeably
- Offer an easier way to choose which algorithm will be used at runtime

#### Image rendering

We want to render images in different formats (PNG and JPEG). The rendering strategy can be swapped at runtime.

#### Acceptance criteria

1. Have a `Renderer` interface with a `Render(data []byte) ([]byte, error)` method.
2. Implement PNG and JPEG renderers.
3. Pick the strategy at runtime based on some condition.

#### Implementation

```go
package behavioral

import "image"

type ImageStrategy interface {
    Draw(canvas []Square) error
    Recolor(square *Square, c color)
}

type Square struct {
    X int
    Y int
}

type color byte

const (
    black color = iota
    white
)

type PNGRenderer struct{}

func (p *PNGRenderer) Draw(canvas []Square) error {
    // PNG rendering implementation
    return nil
}

func (p *PNGRenderer) Recolor(square *Square, c color) {
    // PNG recolor implementation
}
```

### Chain of Responsibility design pattern

#### Description

The Chain of Responsibility pattern passes a request along a chain of handlers. Upon receiving a request, each handler decides either to process the request or to pass it to the next handler in the chain.

#### Objectives

- Decouple the sender of the request from its potential receivers
- Each handler decides to process or pass the request on
- Allow dynamically adding new handlers

#### An event logger chain

We create a logging chain where each logger has a priority. If a log entry's priority is equal to or less than the handler's priority, the handler writes the log; otherwise it passes the message to the next handler.

#### Acceptance criteria

1. A `WithLogger` type must implement the `Handler` interface.
2. There must be three concrete loggers: one for standard output, one for errors, and one that logs to a file.
3. Each logger checks its own priority before writing.

#### Implementation

```go
package behavioral

import "fmt"

const (
    None = iota
    Notice
    Warning
    Critical
)

type LoggingChain struct{}

type ChainLogger interface {
    Next(LoggingChain)
}

type FirstLogger struct {
    LogLevel int
    NextChain ChainLogger
}

func (f *FirstLogger) Println(s string, level int) {
    if f.LogLevel >= level {
        fmt.Printf("[First Logger] %s\n", s)
    } else if f.NextChain != nil {
        // pass to next
    }
}
```

### Command design pattern

#### Description

The Command pattern encapsulates a request as an object, letting you parameterize clients with different requests, queue or log requests, and support undoable operations.

#### Objectives

- Put a layer between the thing that creates the command and the thing that executes it
- Allow implementing undo/redo operations
- Allow queuing of commands

#### A simple queue of commands

We want to create a queue of console commands that can be executed and undone.

#### Acceptance criteria

1. Create a `Command` interface with `Execute()` and `Undo()` methods.
2. Create concrete commands for printing to the console.
3. Use a queue to batch and execute commands.

#### Implementation

```go
package behavioral

import "fmt"

type Command interface {
    Execute() string
}

type ConsoleOutput struct {
    message string
}

func (c *ConsoleOutput) Execute() string {
    return c.message
}

type CommandQueue struct {
    queue []Command
}

func (p *CommandQueue) AddCommand(c Command) {
    p.queue = append(p.queue, c)
    if len(p.queue) == 3 {
        for _, command := range p.queue {
            fmt.Print(command.Execute())
        }
        p.queue = make([]Command, 3)
    }
}
```

---

## Chapter 6. Behavioral Patterns — Template, Memento, and Interpreter

### Template design pattern

#### Description

The Template pattern defines the skeleton of an algorithm in a base method, deferring some steps to subclasses. It lets subclasses redefine certain steps without changing the algorithm's structure.

#### Objectives

- Defer the implementation of some parts of an algorithm to subclasses
- Achieve better code reuse by using the template to define a series of steps

#### A simple retriever

We want a common template for retrieving content from local or remote sources.

#### Acceptance criteria

1. The template must have a `Steps()` method that defines the sequence.
2. Each step can be overridden by the concrete implementation.
3. The template must define a shared step (`MessageRetrieval`) that is not overridden.

#### Implementation

```go
package behavioral

import "fmt"

type Template interface {
    first() int
    second() string
    third() int
}

func RunSteps(t Template) int {
    firstResult := t.first()
    secondResult := t.second()
    thirdResult := t.third()
    fmt.Printf("%d, %s, %d\n", firstResult, secondResult, thirdResult)
    return firstResult + thirdResult
}
```

### Memento design pattern

#### Description

The Memento pattern provides the ability to restore an object to its previous state (undo).

#### Objectives

- Capture an object's internal state
- Store the state in an external object (the Memento)
- Restore the state at a later time

#### A game save state

We want to save and restore the state of a game character.

#### Acceptance criteria

1. Create a `GameCharacter` type with health, attack, and defense.
2. The character must be able to save its state to a `Memento`.
3. The character must be able to restore its state from a `Memento`.

#### Implementation

```go
package behavioral

type Memento struct {
    State string
}

type GameCharacter struct {
    HP      int
    Attack  int
    Defense int
}

func (g *GameCharacter) Save() Memento {
    return Memento{State: fmt.Sprintf("%d:%d:%d", g.HP, g.Attack, g.Defense)}
}

func (g *GameCharacter) Restore(m Memento) {
    // parse and restore state
}
```

### Interpreter design pattern

#### Description

The Interpreter pattern defines a grammar for a language and provides an interpreter to deal with that grammar.

#### Objectives

- Provide grammar for simple languages
- Build a tree of expressions that can be evaluated
- Build a very flexible interpreter that is easy to extend with new expression types

#### A simple math expression parser

We want to evaluate arithmetic expressions like `3 + 4`.

#### Acceptance criteria

1. Each token in an expression must implement an `Interpreter` interface.
2. Provide sum and multiply tokens.
3. Compose tokens into a tree structure.

#### Implementation

```go
package behavioral

type Interpreter interface {
    Read() int
}

type TerminalExpression struct {
    val int
}

func (t *TerminalExpression) Read() int {
    return t.val
}

type NonTerminalExpression struct {
    left  Interpreter
    right Interpreter
}

func (n *NonTerminalExpression) Read() int {
    return n.left.Read() + n.right.Read()
}
```

---

## Chapter 7. Behavioral Patterns — Visitor, State, Mediator, and Observer

### Visitor design pattern

#### Description

The Visitor pattern lets you separate algorithms from the objects on which they operate. You can add further operations to objects without modifying them.

#### Objectives

- Separate an algorithm from the object structure it operates on
- Allow adding new operations without changing the types

#### Calculating prices and printing product lists

We have products (Rice, Pasta, Fridge) that need to be processed by visitors for price calculation and name printing.

#### Acceptance criteria

1. Create a `Visitor` interface with a `Visit(Visitable)` method.
2. Create a `Visitable` interface with an `Accept(Visitor)` method.
3. Implement `PriceVisitor` and `NamePrinter` visitors.

#### Implementation

```go
package behavioral

import "fmt"

type Visitor interface {
    Visit(Visitable)
}

type Visitable interface {
    Accept(Visitor)
}

type ProductInfoRetriever interface {
    GetPrice() float32
    GetName() string
}

type Product struct {
    Price float32
    Name  string
}

func (p *Product) GetPrice() float32 {
    return p.Price
}

func (p *Product) GetName() string {
    return p.Name
}

type Rice struct {
    Product
}

func (r *Rice) Accept(v Visitor) {
    v.Visit(r)
}

type Pasta struct {
    Product
}

func (p *Pasta) Accept(v Visitor) {
    v.Visit(p)
}

type Fridge struct {
    Product
}

func (f *Fridge) GetPrice() float32 {
    return f.Product.Price + 20
}

func (f *Fridge) Accept(v Visitor) {
    v.Visit(f)
}

type PriceVisitor struct {
    Sum float32
}

func (p *PriceVisitor) Visit(v Visitable) {
    if item, ok := v.(ProductInfoRetriever); ok {
        p.Sum += item.GetPrice()
    }
}

type NamePrinter struct {
    ProductList string
}

func (n *NamePrinter) Visit(v Visitable) {
    if item, ok := v.(ProductInfoRetriever); ok {
        n.ProductList = fmt.Sprintf("%s\n%s", item.GetName(), n.ProductList)
    }
}
```

Running this:

```bash
go run visitor.go
```

```
Total: 72.000000
Product list:

Some pasta
Some rice
```

With the Fridge added (price 50 + 20 = 70):

```
Total: 142.000000
Product list:

A fridge
Some pasta
Some rice
```

### State design pattern

#### Description

The State pattern lets an object alter its behavior when its internal state changes. The object will appear to change its class.

#### Objectives

- Have a type that alters its own behavior when some internal things have changed
- Model complex graphs and pipelines that can be upgraded easily by adding more states

#### A number-guessing game

We implement a simple number-guessing game using a Finite State Machine (FSM). The player chooses a difficulty level (number of retries), then tries to guess a number between 0 and 10.

#### Acceptance criteria

1. The game asks the player how many tries they will have.
2. The number to guess must be between 0 and 10.
3. Every time a player enters a number, retries drop by one.
4. If retries reach zero and the guess is incorrect, the player loses.
5. If the player guesses the number, the player wins.

#### Implementation

```go
package behavioral

import (
    "fmt"
    "math/rand"
    "os"
    "time"
)

type GameState interface {
    executeState(*GameContext) bool
}

type GameContext struct {
    SecretNumber int
    Retries      int
    Won          bool
    Next         GameState
}

type StartState struct{}

func (s *StartState) executeState(c *GameContext) bool {
    c.Next = &AskState{}
    rand.Seed(time.Now().UnixNano())
    c.SecretNumber = rand.Intn(10)
    fmt.Println("Introduce a number a number of retries to set the difficulty:")
    fmt.Fscanf(os.Stdin, "%d\n", &c.Retries)
    return true
}

type AskState struct{}

func (a *AskState) executeState(c *GameContext) bool {
    fmt.Printf("Introduce a number between 0 and 10, you have %d tries left\n", c.Retries)
    var n int
    fmt.Fscanf(os.Stdin, "%d", &n)
    c.Retries = c.Retries - 1
    if n == c.SecretNumber {
        c.Won = true
        c.Next = &FinishState{}
    }
    if c.Retries == 0 {
        c.Next = &FinishState{}
    }
    return true
}

type FinishState struct{}

func (f *FinishState) executeState(c *GameContext) bool {
    if c.Won {
        c.Next = &WinState{}
    } else {
        c.Next = &LoseState{}
    }
    return true
}

type WinState struct{}

func (w *WinState) executeState(c *GameContext) bool {
    println("Congrats, you won")
    return false
}

type LoseState struct{}

func (l *LoseState) executeState(c *GameContext) bool {
    fmt.Printf("You lose. The correct number was: %d\n", c.SecretNumber)
    return false
}

func main() {
    start := StartState{}
    game := GameContext{
        Next: &start,
    }
    for game.Next.executeState(&game) {}
}
```

### Mediator design pattern

#### Description

The Mediator pattern defines an object that encapsulates how a set of objects interact. It promotes loose coupling by keeping objects from referring to each other explicitly.

#### Objectives

- Provide loose coupling between two objects that must communicate
- Reduce the amount of dependencies a particular type has by passing these needs to the Mediator

#### A calculator

Rather than having every numeric type know how to interact with every other numeric type, a `Sum` mediator function centralizes the knowledge:

```go
func Sum(a, b interface{}) interface{} {
    switch a := a.(type) {
    case One:
        switch b.(type) {
        case One:
            return &Two{}
        case Two:
            return &Three{}
        default:
            return fmt.Errorf("Number not found")
        }
    case int:
        switch b := b.(type) {
        case int:
            return a + b
        default:
            return fmt.Errorf("Number not found")
        }
    default:
        return fmt.Errorf("Number not found")
    }
}
```

Running:

```bash
go run mediator.go
```

```
&main.Three{}
3
```

The Mediator knows about all possible types and returns the most convenient result. Without it, each type would need to implement operations against every other type, creating extreme coupling.

### Observer design pattern

#### Description

The Observer pattern (also known as Publish/Subscribe) defines a one-to-many dependency between objects so that when one object changes state, all its dependents are notified automatically.

#### Objectives

- Provide an event-driven architecture where one event can trigger one or more actions
- Uncouple the actions that are performed from the event that triggers them
- Provide more than one event that triggers the same action

#### The notifier

We build a `Publisher` that maintains a list of observers and notifies them when events occur.

#### Acceptance criteria

1. A publisher with a `NotifyObservers` method that accepts a message and triggers `Notify` on every subscribed observer.
2. A method to add new subscribers.
3. A method to remove subscribers.

#### Unit tests

```go
type TestObserver struct {
    ID      int
    Message string
}

func (p *TestObserver) Notify(m string) {
    fmt.Printf("Observer %d: message '%s' received \n", p.ID, m)
    p.Message = m
}

func TestSubject(t *testing.T) {
    testObserver1 := &TestObserver{1, ""}
    testObserver2 := &TestObserver{2, ""}
    testObserver3 := &TestObserver{3, ""}
    publisher := Publisher{}

    t.Run("AddObserver", func(t *testing.T) {
        publisher.AddObserver(testObserver1)
        publisher.AddObserver(testObserver2)
        publisher.AddObserver(testObserver3)
        if len(publisher.ObserversList) != 3 {
            t.Fail()
        }
    })

    t.Run("RemoveObserver", func(t *testing.T) {
        publisher.RemoveObserver(testObserver2)
        if len(publisher.ObserversList) != 2 {
            t.Errorf("The size of the observer list is not the expected. 3 != %d\n",
                len(publisher.ObserversList))
        }
        for _, observer := range publisher.ObserversList {
            testObserver, ok := observer.(*TestObserver)
            if !ok {
                t.Fail()
            }
            if testObserver.ID == 2 {
                t.Fail()
            }
        }
    })

    t.Run("Notify", func(t *testing.T) {
        if len(publisher.ObserversList) == 0 {
            t.Errorf("The list is empty. Nothing to test\n")
        }
        message := "Hello World!"
        publisher.NotifyObservers(message)
        for _, observer := range publisher.ObserversList {
            printObserver, ok := observer.(*TestObserver)
            if !ok {
                t.Fail()
                break
            }
            if printObserver.Message != message {
                t.Errorf("Expected message on observer %d was not expected: '%s' != '%s'\n",
                    printObserver.ID, printObserver.Message, message)
            }
        }
    })
}
```

#### Implementation

```go
type Observer interface {
    Notify(string)
}

type Publisher struct {
    ObserversList []Observer
}

func (s *Publisher) AddObserver(o Observer) {
    s.ObserversList = append(s.ObserversList, o)
}

func (s *Publisher) RemoveObserver(o Observer) {
    var indexToRemove int
    for i, observer := range s.ObserversList {
        if observer == o {
            indexToRemove = i
            break
        }
    }
    s.ObserversList = append(s.ObserversList[:indexToRemove],
        s.ObserversList[indexToRemove+1:]...)
}

func (s *Publisher) NotifyObservers(m string) {
    fmt.Printf("Publisher received message '%s' to notify observers\n", m)
    for _, observer := range s.ObserversList {
        observer.Notify(m)
    }
}
```

Test output:

```
Publisher received message 'Hello World!' to notify observers
Observer 1: message 'Hello World!' received
Observer 3: message 'Hello World!' received
```

---

## Chapter 8. Introduction to Go's Concurrency

### A little bit of history and theory

CPUs improved in speed until reaching hardware limits, then multicore processors emerged. Languages like Java and C++ were designed for single-core CPUs and require third-party tools for robust concurrency. Go was designed with concurrency as a first-class concern.

### Concurrency versus parallelism

- **Concurrency** is about dealing with many things at once (structure)
- **Parallelism** is about doing many things at the same time (execution)

Concurrency enables parallelism by designing correct concurrent structure. Think of a bicycle: two legs push pedals concurrently (concurrent design). A tandem bicycle with two riders is both concurrent and parallel.

The key insight: with a concurrent design, you don't need to worry about parallelism. Parallelism is an automatic bonus if the hardware supports it.

You can control the number of CPU cores Go uses:

```bash
GOMAXPROCS=4 go run main.go
```

### CSP versus actor-based concurrency

The most common concurrency model is the Actor model: actors communicate by sending messages directly to each other by actor ID.

Go uses **Communicating Sequential Processes (CSP)**: processes are anonymous, and communication happens through named channels. Neither the sender nor receiver needs to know about the other — they only need to know the channel.

### Goroutines

Goroutines are Go's concurrent execution units. They are extremely lightweight (much cheaper than OS threads). Millions of Goroutines can run simultaneously.

#### Our first Goroutine

```go
package main

func main() {
    helloWorld()
}

func helloWorld() {
    println("Hello World!")
}
```

To run it in a new Goroutine:

```go
package main

func main() {
    go helloWorld()
}

func helloWorld() {
    println("Hello World!")
}
```

Running this prints nothing because `main` finishes before the Goroutine executes.

#### WaitGroups

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wait sync.WaitGroup
    wait.Add(1)

    go func() {
        fmt.Println("Hello World!")
        wait.Done()
    }()

    wait.Wait()
}
```

Launching multiple Goroutines with a WaitGroup:

```go
func main() {
    var wait sync.WaitGroup
    goRoutines := 5
    wait.Add(goRoutines)

    for i := 0; i < goRoutines; i++ {
        go func(goRoutineID int) {
            fmt.Printf("ID:%d: Hello goroutines!\n", goRoutineID)
            wait.Done()
        }(i)
    }
    wait.Wait()
}
```

Output (order is non-deterministic):

```
ID:4: Hello goroutines!
ID:0: Hello goroutines!
ID:1: Hello goroutines!
ID:2: Hello goroutines!
ID:3: Hello goroutines!
```

### Callbacks

A callback is an anonymous function executed within the context of another function:

```go
func toUpperSync(word string, f func(string)) {
    f(strings.ToUpper(word))
}

func main() {
    toUpperSync("Hello Callbacks!", func(v string) {
        fmt.Printf("Callback: %s\n", v)
    })
}
```

Asynchronous callback:

```go
var wait sync.WaitGroup

func toUpperAsync(word string, f func(string)) {
    go func() {
        f(strings.ToUpper(word))
    }()
}

func main() {
    wait.Add(1)
    toUpperAsync("Hello Callbacks!", func(v string) {
        fmt.Printf("Callback: %s\n", v)
        wait.Done()
    })
    println("Waiting async response...")
    wait.Wait()
}
```

Output:

```
Waiting async response...
Callback: HELLO CALLBACKS!
```

#### Callback hell

Nesting too many callbacks becomes hard to reason about:

```go
toUpperAsync("Hello Callbacks!", func(v string) {
    toUpperAsync(fmt.Sprintf("Callback: %s\n", v), func(v string) {
        fmt.Printf("Callback within %s", v)
        wait.Done()
    })
})
```

### Mutexes

A mutex controls exclusive access to a shared resource. It prevents race conditions.

```go
type Counter struct {
    sync.Mutex
    value int
}

func main() {
    counter := Counter{}
    for i := 0; i < 10; i++ {
        go func(i int) {
            counter.Lock()
            counter.value++
            defer counter.Unlock()
        }(i)
    }
    time.Sleep(time.Second)
    counter.Lock()
    defer counter.Unlock()
    println(counter.value)
}
```

### Presenting the race detector

Go includes a built-in race detector:

```bash
go run -race main.go
```

With a race condition:

```
WARNING: DATA RACE
Read at 0x00c42007a068 by goroutine 6:
...
Previous write at 0x00c42007a068 by goroutine 5:
...
Found 1 data race(s)
exit status 66
```

> **Note:** The race detector works at runtime, not statically. It only catches races that actually occur during execution.

### Channels

Channels allow communication between Goroutines. They are the idiomatic Go way to synchronize concurrent code.

#### Our first channel

```go
package main

import "fmt"

func main() {
    channel := make(chan string)
    go func() {
        channel <- "Hello World!"
    }()
    message := <-channel
    fmt.Println(message)
}
```

Channels are blocking by default: a sender blocks until a receiver takes the value, and vice versa.

#### Buffered channels

```go
channel := make(chan string, 1)
```

A buffered channel with capacity 1 does not block the sender until the buffer is full.

#### Directional channels

Restrict a channel to send-only or receive-only:

```go
// Send-only channel
go func(ch chan<- string) {
    ch <- "Hello World!"
}(channel)

// Receive-only channel
func receivingCh(ch <-chan string) {
    msg := <-ch
    println(msg)
}
```

#### The select statement

Handle multiple channels in one Goroutine:

```go
func receiver(helloCh, goodbyeCh <-chan string, quitCh chan<- bool) {
    for {
        select {
        case msg := <-helloCh:
            println(msg)
        case msg := <-goodbyeCh:
            println(msg)
        case <-time.After(time.Second * 2):
            println("Nothing received in 2 seconds. Exiting")
            quitCh <- true
            return
        }
    }
}
```

#### Ranging over channels

```go
for v := range ch {
    println(v)
}
```

The range loop continues until the channel is closed.

### Using it all — concurrent singleton

A concurrent counter using channels:

```go
var addCh chan bool = make(chan bool)
var getCountCh chan chan int = make(chan chan int)
var quitCh chan bool = make(chan bool)

func init() {
    var count int
    go func(addCh <-chan bool, getCountCh <-chan chan int, quitCh <-chan bool) {
        for {
            select {
            case <-addCh:
                count++
            case ch := <-getCountCh:
                ch <- count
            case <-quitCh:
                return
            }
        }
    }(addCh, getCountCh, quitCh)
}

type singleton struct{}

var instance singleton

func GetInstance() *singleton {
    return &instance
}

func (s *singleton) AddOne() {
    addCh <- true
}

func (s *singleton) GetCount() int {
    resCh := make(chan int)
    defer close(resCh)
    getCountCh <- resCh
    return <-resCh
}

func (s *singleton) Stop() {
    quitCh <- true
    close(addCh)
    close(getCountCh)
    close(quitCh)
}
```

Alternatively, use `sync.RWMutex`:

```go
type singleton struct {
    count int
    sync.RWMutex
}

func (s *singleton) AddOne() {
    s.Lock()
    defer s.Unlock()
    s.count++
}

func (s *singleton) GetCount() int {
    s.RLock()
    defer s.RUnlock()
    return s.count
}
```

`sync.RWMutex` allows multiple concurrent readers but only one writer at a time.

---

## Chapter 9. Concurrency Patterns — Barrier, Future, and Pipeline

### Barrier concurrency pattern

#### Description

The Barrier pattern blocks execution until all results are ready. It is useful when you need to compose a response from multiple concurrent operations.

#### Objectives

- Compose the value of a type with the data coming from one or more Goroutines
- Control the correctness of any of those incoming data pipes so that no inconsistent data is returned

#### An HTTP GET aggregator

We perform two HTTP GET calls concurrently and print a composed response only when both succeed.

#### Acceptance criteria

- Print the merged result of two calls to `http://httpbin.org/headers` and `http://httpbin.org/User-Agent`.
- If any call fails, print only the error message.
- Print the result as a composed response when both calls have finished.

#### Implementation

```go
package barrier

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

var timeoutMilliseconds int = 5000

type barrierResp struct {
    Err  error
    Resp string
}

func barrier(endpoints ...string) {
    requestNumber := len(endpoints)
    in := make(chan barrierResp, requestNumber)
    defer close(in)
    responses := make([]barrierResp, requestNumber)

    for _, endpoint := range endpoints {
        go makeRequest(in, endpoint)
    }

    var hasError bool
    for i := 0; i < requestNumber; i++ {
        resp := <-in
        if resp.Err != nil {
            fmt.Println("ERROR: ", resp.Err)
            hasError = true
        }
        responses[i] = resp
    }

    if !hasError {
        for _, resp := range responses {
            fmt.Println(resp.Resp)
        }
    }
}

func makeRequest(out chan<- barrierResp, url string) {
    res := barrierResp{}
    client := http.Client{
        Timeout: time.Duration(time.Duration(timeoutMilliseconds) * time.Millisecond),
    }
    resp, err := client.Get(url)
    if err != nil {
        res.Err = err
        out <- res
        return
    }
    byt, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        res.Err = err
        out <- res
        return
    }
    res.Resp = string(byt)
    out <- res
}
```

> **Note:** In testing, watch out for false positives (a test that passes when it should fail) and false negatives (a test that fails when behavior is actually correct). Always verify your tests actually fail before implementing.

### Future design pattern

#### Description

The Future pattern (also called Promise) enables asynchronous programming by pre-defining success and failure handlers before executing an operation.

#### Objectives

- Delegate the action handler to a different Goroutine
- Stack many asynchronous calls between them

#### A simple asynchronous requester

We define a `MaybeString` type that chains success and failure handlers before executing a function.

#### Implementation

```go
package future

type SuccessFunc func(string)
type FailFunc func(error)
type ExecuteStringFunc func() (string, error)

type MaybeString struct {
    successFunc SuccessFunc
    failFunc    FailFunc
}

func (s *MaybeString) Success(f SuccessFunc) *MaybeString {
    s.successFunc = f
    return s
}

func (s *MaybeString) Fail(f FailFunc) *MaybeString {
    s.failFunc = f
    return s
}

func (s *MaybeString) Execute(f ExecuteStringFunc) {
    go func(s *MaybeString) {
        str, err := f()
        if err != nil {
            s.failFunc(err)
        } else {
            s.successFunc(str)
        }
    }(s)
}
```

Usage with closures:

```go
func setContext(msg string) ExecuteStringFunc {
    msg = fmt.Sprintf("%s Closure!\n", msg)
    return func() (string, error) {
        return msg, nil
    }
}

future.Success(func(s string) {
    fmt.Println(s)
}).Fail(func(e error) {
    fmt.Println("Error:", e)
})
future.Execute(setContext("Hello"))
```

Output:

```
Hello Closure!
```

### Pipeline design pattern

#### Description

The Pipeline pattern connects Goroutines with channels to form a multi-stage processing chain. Each stage takes values from an input channel and sends results to an output channel.

#### Objectives

- Create a concurrent structure of a multistep algorithm
- Exploit the parallelism of multicore machines by decomposing an algorithm into different Goroutines

#### A concurrent multi-operation

We generate numbers 1..N, square them, then sum the results.

For N=3: [1,2,3] → [1,4,9] → sum=14
For N=5: [1,2,3,4,5] → [1,4,9,16,25] → sum=55

#### Implementation

Each pipeline step follows this pattern: create a channel, launch a Goroutine, return the channel immediately:

```go
package pipelines

func generator(max int) <-chan int {
    outChInt := make(chan int, 100)
    go func() {
        for i := 1; i <= max; i++ {
            outChInt <- i
        }
        close(outChInt)
    }()
    return outChInt
}

func power(in <-chan int) <-chan int {
    out := make(chan int, 100)
    go func() {
        for v := range in {
            out <- v * v
        }
        close(out)
    }()
    return out
}

func sum(in <-chan int) <-chan int {
    out := make(chan int, 100)
    go func() {
        var sum int
        for v := range in {
            sum += v
        }
        out <- sum
        close(out)
    }()
    return out
}

func LaunchPipeline(amount int) int {
    return <-sum(power(generator(amount)))
}
```

> **Note:** The `for v := range ch` loop keeps taking values from a channel indefinitely until the channel is closed. Always close channels when done sending to prevent Goroutine leaks.

Test output:

```
14 == 14
55 == 55
```

---

## Chapter 10. Concurrency Patterns — Workers Pool and Publish/Subscriber

### Workers pool

#### Description

A workers pool bounds the number of Goroutines to control resource usage. All workers share a single input channel; when work arrives, whichever worker is free picks it up.

#### Objectives

- Control access to shared resources using quotas
- Create a limited amount of Goroutines per app
- Provide more parallelism capabilities to other concurrent structures

#### A pool of pipelines

We create a pipeline that uppercases a string, appends a suffix, and prefixes a worker ID.

#### Acceptance criteria

1. When making a request with a string value, it must be uppercased.
2. Once uppercase, a predefined text must be appended (not uppercase).
3. The worker ID must be prefixed to the final string.
4. The resulting string must be passed to a predefined handler.

#### Implementation

```go
// workers_pipeline.go

type Request struct {
    Data    interface{}
    Handler RequestHandler
}

type RequestHandler func(interface{})

type WorkerLauncher interface {
    LaunchWorker(in chan Request)
}

type Dispatcher interface {
    LaunchWorker(w WorkerLauncher)
    MakeRequest(Request)
    Stop()
}

type dispatcher struct {
    inCh chan Request
}

func (d *dispatcher) LaunchWorker(w WorkerLauncher) {
    w.LaunchWorker(d.inCh)
}

func (d *dispatcher) Stop() {
    close(d.inCh)
}

func (d *dispatcher) MakeRequest(r Request) {
    select {
    case d.inCh <- r:
    case <-time.After(time.Second * 5):
        return
    }
}

func NewDispatcher(b int) Dispatcher {
    return &dispatcher{
        inCh: make(chan Request, b),
    }
}

// worker.go

type PreffixSuffixWorker struct {
    id      int
    prefixS string
    suffixS string
}

func (w *PreffixSuffixWorker) uppercase(in <-chan Request) <-chan Request {
    out := make(chan Request)
    go func() {
        for msg := range in {
            s, ok := msg.Data.(string)
            if !ok {
                msg.Handler(nil)
                continue
            }
            msg.Data = strings.ToUpper(s)
            out <- msg
        }
        close(out)
    }()
    return out
}

func (w *PreffixSuffixWorker) append(in <-chan Request) <-chan Request {
    out := make(chan Request)
    go func() {
        for msg := range in {
            uppercaseString, ok := msg.Data.(string)
            if !ok {
                msg.Handler(nil)
                continue
            }
            msg.Data = fmt.Sprintf("%s%s", uppercaseString, w.suffixS)
            out <- msg
        }
        close(out)
    }()
    return out
}

func (w *PreffixSuffixWorker) prefix(in <-chan Request) {
    go func() {
        for msg := range in {
            uppercasedStringWithSuffix, ok := msg.Data.(string)
            if !ok {
                msg.Handler(nil)
                continue
            }
            msg.Handler(fmt.Sprintf("%s%s", w.prefixS, uppercasedStringWithSuffix))
        }
    }()
}

func (w *PreffixSuffixWorker) LaunchWorker(in chan Request) {
    w.prefix(w.append(w.uppercase(in)))
}
```

Running the app:

```bash
go run *
```

```
WorkerID: 1 -> (MSG_ID: 0) -> HELLO World
WorkerID: 0 -> (MSG_ID: 3) -> HELLO World
WorkerID: 2 -> (MSG_ID: 4) -> HELLO World
...
```

The pipeline stages execute in order: uppercase → append suffix → prefix worker ID. Requests are distributed across workers non-deterministically.

Stopping the dispatcher closes the shared input channel, which propagates through all pipeline stages via for-range loops, gracefully terminating all Goroutines.

### Concurrent Publish/Subscriber design pattern

#### Description

This is a concurrent reimplementation of the Observer pattern from Chapter 7. Each subscriber runs in its own Goroutine. Access to the subscriber list is serialized through channels to avoid race conditions.

#### New challenges compared to the non-concurrent Observer

- Access to the subscriber list must be serialized (use channels, not mutexes)
- When a subscriber is removed, its Goroutine must be stopped
- When the publisher stops, all subscriber Goroutines must stop too

#### Objectives

- Provide an event-driven architecture where one event can trigger one or more actions
- Uncouple the actions from the events that trigger them
- All inter-Goroutine communication must be synchronized with timeouts

#### Subscriber interface

```go
type Subscriber interface {
    Notify(interface{}) error
    Close()
}
```

#### Publisher interface

```go
type Publisher interface {
    start()
    AddSubscriberCh() chan<- Subscriber
    RemoveSubscriberCh() chan<- Subscriber
    PublishingCh() chan<- interface{}
    Stop()
}
```

#### writerSubscriber implementation

```go
type writerSubscriber struct {
    in     chan interface{}
    id     int
    Writer io.Writer
}

func NewWriterSubscriber(id int, out io.Writer) Subscriber {
    if out == nil {
        out = os.Stdout
    }
    s := &writerSubscriber{
        id:     id,
        in:     make(chan interface{}),
        Writer: out,
    }
    go func() {
        for msg := range s.in {
            fmt.Fprintf(s.Writer, "(W%d): %v\n", s.id, msg)
        }
    }()
    return s
}

func (s *writerSubscriber) Close() {
    close(s.in)
}

func (s *writerSubscriber) Notify(msg interface{}) (err error) {
    defer func() {
        if rec := recover(); rec != nil {
            err = fmt.Errorf("%#v", rec)
        }
    }()
    select {
    case s.in <- msg:
    case <-time.After(time.Second):
        err = fmt.Errorf("Timeout\n")
    }
    return
}
```

The `Notify` method uses two protection mechanisms:
1. A deferred `recover()` to catch panics from sending on a closed channel.
2. A `select` with a 1-second timeout to prevent blocking forever.

#### publisher implementation

```go
type publisher struct {
    subscribers []Subscriber
    addSubCh    chan Subscriber
    removeSubCh chan Subscriber
    in          chan interface{}
    stop        chan struct{}
}

func NewPublisher() Publisher {
    return &publisher{}
}

func (p *publisher) AddSubscriberCh() chan<- Subscriber {
    return p.addSubCh
}

func (p *publisher) RemoveSubscriberCh() chan<- Subscriber {
    return p.removeSubCh
}

func (p *publisher) PublishingCh() chan<- interface{} {
    return p.in
}

func (p *publisher) Stop() {
    close(p.stop)
}

func (p *publisher) start() {
    for {
        select {
        case msg := <-p.in:
            for _, sub := range p.subscribers {
                sub.Notify(msg)
            }
        case sub := <-p.addSubCh:
            p.subscribers = append(p.subscribers, sub)
        case sub := <-p.removeSubCh:
            for i, candidate := range p.subscribers {
                if candidate == sub {
                    p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
                    candidate.Close()
                    break
                }
            }
        case <-p.stop:
            for _, sub := range p.subscribers {
                sub.Close()
            }
            close(p.addSubCh)
            close(p.in)
            close(p.removeSubCh)
            return
        }
    }
}
```

The `start` method's `select` statement serializes all access to the subscriber list:
- Only one case executes per iteration — no two cases can run simultaneously
- This eliminates the need for mutexes when accessing `p.subscribers`
- Closing `p.stop` triggers cleanup of all subscriber Goroutines and closes all channels

Running all tests:

```bash
go test -race .
ok
```

#### Summary of the concurrent Observer

The channel-based approach is more verbose than a mutex-based approach, but it demonstrates the power of Go's concurrency model. Every dangerous operation on shared state is protected by being routed through the publisher's `start` Goroutine via channels.

---

## Summary

This book covered all major Gang of Four design patterns adapted to Go, plus concurrency patterns specific to Go's CSP model:

**Creational patterns:** Singleton, Builder, Factory, Abstract Factory, Prototype

**Structural patterns:** Composite, Adapter, Bridge, Proxy, Facade, Decorator, Flyweight

**Behavioral patterns:** Strategy, Chain of Responsibility, Command, Template, Memento, Interpreter, Visitor, State, Mediator, Observer

**Concurrency patterns:** Barrier, Future, Pipeline, Workers Pool, Concurrent Publish/Subscribe

Key Go concurrency takeaways:
- Think in terms of concurrent structure, not parallel execution
- Channels are the idiomatic way to communicate between Goroutines
- The `select` statement serializes access to multiple channels
- Always close channels when done sending; for-range loops stop on channel close
- Use `go test -race` to detect race conditions at runtime
- WaitGroups synchronize Goroutine completion
- The `sync.RWMutex` allows concurrent reads but exclusive writes
