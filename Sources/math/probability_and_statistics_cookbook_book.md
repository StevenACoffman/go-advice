# Probability and Statistics Cookbook

*Matthias Vallentin, 2011*

This cookbook integrates a variety of topics in probability theory and statistics. It is based on literature [1, 6, 3] and in-class
material from courses of the statistics department at the Uni-

versity of California in Berkeley but also influenced by other
sources [4, 5]. If you find errors or have suggestions for further
topics, I would appreciate if you send me an email. The most recent version of this document is available at http://matthias.
vallentin.net/probability-and-statistics-cookbook/.
To
reproduce, please contact me.

## 1 Distribution Overview

### 1.1 Discrete Distributions

Notation1
FX(x)
fX(x)
E [X]
V [X]
MX(s)

Uniform
Unif {a, . . . , b}

0
x < a
⌊x⌋-a+1

b-a
a ≤x ≤b
1
x > b

I(a < x < b)

b -a + 1
a + b

2
(b -a + 1)2 -1

12
eas -e-(b+1)s

s(b -a)

Bernoulli
Bern (p)
(1 -p)1-x
px (1 -p)1-x
p
p(1 -p)
1 -p + pes

Binomial
Bin (n, p)
I1-p(n -x, x + 1)

�
n
x

�

px (1 -p)n-x
np
np(1 -p)
(1 -p + pes)n

Multinomial
Mult (n, p)
n!
x1! . . . xk!px1
1 · · · pxk
k

k
�

i=1
xi = n
npi
npi(1 -pi)

�k
�

i=0
piesi
�n

Hypergeometric
Hyp (N, m, n)
≈Φ

�
x -np
�
np(1 -p)

�
�m
x
��m-x
n-x
�

�N
x
�
nm

N
nm(N -n)(N -m)

N 2(N -1)

Negative Binomial
NBin (r, p)
Ip(r, x + 1)

�
x + r -1
r -1

�

pr(1 -p)x
r 1 -p

p
r 1 -p

p2

�
p
1 -(1 -p)es

�r

Geometric
Geo (p)
1 -(1 -p)x
x ∈N+
p(1 -p)x-1
x ∈N+
1
p
1 -p

p2
p
1 -(1 -p)es

Poisson
Po (λ)
e-λ
x
�

i=0

λi

i!
λxe-λ

x!
λ
λ
eλ(es-1)

G
G
G
G
G
G

Uniform (discrete)

x

PMF

a
b

1

n

G G G G

G

G

G

G

G

G

G

G

G

G

G

G

G

G

G

G G G G G G G G G G G G G G G G G G G G G

0
10
20
30
40

0.00
0.05
0.10
0.15
0.20
0.25

Binomial

x

PMF

G
n = 40, p = 0.3
n = 30, p = 0.6
n = 25, p = 0.9

G

G

G

G

G

G

G
G
G
G

0
2
4
6
8
10

0.0
0.2
0.4
0.6
0.8

Geometric

x

PMF

G
p = 0.2
p = 0.5
p = 0.8

G
G

G

G

G

G
G
G
G
G
G
G
G
G
G
G
G
G
G
G
G

0
5
10
15
20

0.0
0.1
0.2
0.3

Poisson

x

PMF

G
λ = 1
λ = 4
λ = 10

1We use the notation γ(s, x) and Γ(x) to refer to the Gamma functions (see §22.1), and use B(x, y) and Ix to refer to the Beta functions (see §22.2).

### 1.2 Continuous Distributions

Notation
FX(x)
fX(x)
E [X]
V [X]
MX(s)

Uniform
Unif (a, b)

0
x < a

x-a
b-a
a < x < b
1
x > b

I(a < x < b)

b -a
a + b

2
(b -a)2

12
esb -esa

s(b -a)

Normal
N
�
µ, σ2�
Φ(x) =
�x

-∞
φ(t) dt
φ(x) =
1
σ
√

2π
exp
�
-(x -µ)2

2σ2

�
µ
σ2
exp
�
µs + σ2s2

2

�

Log-Normal
ln N
�
µ, σ2�
1
2 + 1

2 erf
�ln x -µ
√

2σ2

�
1
x
√

2πσ2 exp
�
-(ln x -µ)2

2σ2

�
eµ+σ2/2
(eσ2 -1)e2µ+σ2

Multivariate Normal
MVN (µ, Σ)
(2π)-k/2|Σ|-1/2e-1

2 (x-µ)T Σ-1(x-µ)
µ
Σ
exp
�
µT s + 1

2sT Σs
�

Student's t
Student(ν)
Ix
�ν
2 , ν

2

�
Γ
�ν+1
2
�

√νπΓ
�ν
2
�
�
1 + x2

ν

�-(ν+1)/2
0
0

Chi-square
χ2
k
1
Γ(k/2)γ
�k
2 , x

2

�
1
2k/2Γ(k/2)xk/2-1e-x/2
k
2k
(1 -2s)-k/2 s < 1/2

F
F(d1, d2)
I
d1x
d1x+d2

�d1
2 , d1

2

�
�

(d1x)d1 dd2
2
(d1x+d2)d1+d2

xB
�d1
2 , d1

2
�
d2
d2 -2
2d2
2(d1 + d2 -2)
d1(d2 -2)2(d2 -4)

Exponential
Exp (β)
1 -e-x/β
1
β e-x/β
β
β2
1
1 -βs (s < 1/β)

Gamma
Gamma (α, β)
γ(α, x/β)

Γ(α)
1
Γ (α) βα xα-1e-x/β
αβ
αβ2
�
1
1 -βs

�α
(s < 1/β)

Inverse Gamma
InvGamma (α, β)
Γ
�
α, β

x
�

Γ (α)
βα

Γ (α)x-α-1e-β/x
β
α -1 α > 1
β2

(α -1)2(α -2)2 α > 2
2(-βs)α/2

Γ(α)
Kα
��
-4βs
�

Dirichlet
Dir (α)
Γ
��k
i=1 αi
�

�k
i=1 Γ (αi)

k
�

i=1
xαi-1
i
αi
�k
i=1 αi

E [Xi] (1 -E [Xi])
�k
i=1 αi + 1

Beta
Beta (α, β)
Ix(α, β)
Γ (α + β)
Γ (α) Γ (β)xα-1 (1 -x)β-1
α
α + β
αβ
(α + β)2(α + β + 1)
1 +

∞
�

k=1

�k-1
�

r=0

α + r
α + β + r

�
sk

k!

Weibull
Weibull(λ, k)
1 -e-(x/λ)k
k
λ

�x
λ

�k-1
e-(x/λ)k
λΓ
�
1 + 1

k

�
λ2Γ
�
1 + 2

k

�
-µ2
∞
�

n=0

snλn

n! Γ
�
1 + n

k

�

Pareto
Pareto(xm, α)
1 -
�xm
x

�α
x ≥xm
α xα
m
xα+1
x ≥xm
αxm
α -1 α > 1
xα
m
(α -1)2(α -2) α > 2
α(-xms)αΓ(-α, -xms) s < 0

G
G

Uniform (continuous)

x

PDF

a
b

1

b -a

G
G

-4
-2
0
2
4

0.0
0.2
0.4
0.6
0.8

Normal

x

φ(x)

µ = 0, σ2 = 0.2
µ = 0, σ2 = 1
µ = 0, σ2 = 5
µ = -2, σ2 = 0.5

0.0
0.5
1.0
1.5
2.0
2.5
3.0

0.0
0.2
0.4
0.6
0.8
1.0

Log-normal

x

PDF

µ = 0, σ2 = 3
µ = 2, σ2 = 2
µ = 0, σ2 = 1
µ = 0.5, σ2 = 1
µ = 0.25, σ2 = 1
µ = 0.125, σ2 = 1

-4
-2
0
2
4

0.0
0.1
0.2
0.3
0.4

Student's t

x

PDF

ν = 1
ν = 2
ν = 5
ν = ∞

0
2
4
6
8

0.0
0.1
0.2
0.3
0.4
0.5

χ2

x

PDF

k = 1
k = 2
k = 3
k = 4
k = 5

0
1
2
3
4
5

0.0
0.5
1.0
1.5
2.0
2.5
3.0

F

x

PDF

d1 = 1, d2 = 1
d1 = 2, d2 = 1
d1 = 5, d2 = 2
d1 = 100, d2 = 1
d1 = 100, d2 = 100

0
1
2
3
4
5

0.0
0.5
1.0
1.5
2.0

Exponential

x

PDF

β = 2
β = 1
β = 0.4

0
5
10
15
20

0.0
0.1
0.2
0.3
0.4
0.5

Gamma

x

PDF

α = 1, β = 2
α = 2, β = 2
α = 3, β = 2
α = 5, β = 1
α = 9, β = 0.5

0
1
2
3
4
5

0
1
2
3
4

Inverse Gamma

x

PDF

α = 1, β = 1
α = 2, β = 1
α = 3, β = 1
α = 3, β = 0.5

0.0
0.2
0.4
0.6
0.8
1.0

0.0
0.5
1.0
1.5
2.0
2.5
3.0

Beta

x

PDF

α = 0.5, β = 0.5
α = 5, β = 1
α = 1, β = 3
α = 2, β = 2
α = 2, β = 5

0.0
0.5
1.0
1.5
2.0
2.5

0.0
0.5
1.0
1.5
2.0
2.5

Weibull

x

PDF

λ = 1, k = 0.5
λ = 1, k = 1
λ = 1, k = 1.5
λ = 1, k = 5

0
1
2
3
4
5

0
1
2
3

Pareto

x

PDF

xm = 1, α = 1
xm = 1, α = 2
xm = 1, α = 4

## 2 Probability Theory

Definitions

• Sample space Ω
• Outcome (point or element) ω ∈Ω
• Event A ⊆Ω
• σ-algebra A

1. ∅∈A
2. A1, A2, . . . , ∈A =⇒�∞
i=1 Ai ∈A
3. A ∈A =⇒¬A ∈A

• Probability Distribution P

1. P [A] ≥0
∀A
2. P [Ω] = 1

3. P

�∞
�

i=1
Ai

�

=

∞
�

i=1
P [Ai]

• Probability space (Ω, A, P)

Properties

• P [∅] = 0
• B = Ω∩B = (A ∪¬A) ∩B = (A ∩B) ∪(¬A ∩B)
• P [¬A] = 1 -P [A]
• P [B] = P [A ∩B] + P [¬A ∩B]
• P [Ω] = 1
P [∅] = 0
• ¬(�

n An) = �

n ¬An
¬(�

n An) = �

n ¬An
DeMorgan
• P [�

n An] = 1 -P [�

n ¬An]
• P [A ∪B] = P [A] + P [B] -P [A ∩B]

=⇒P [A ∪B] ≤P [A] + P [B]
• P [A ∪B] = P [A ∩¬B] + P [¬A ∩B] + P [A ∩B]
• P [A ∩¬B] = P [A] -P [A ∩B]

Continuity of Probabilities

• A1 ⊂A2 ⊂. . . =⇒limn→∞P [An] = P [A]
whereA = �∞
i=1 Ai
• A1 ⊃A2 ⊃. . . =⇒limn→∞P [An] = P [A]
whereA = �∞
i=1 Ai

Independence ⊥⊥
A ⊥⊥B ⇐⇒P [A ∩B] = P [A] P [B]

Conditional Probability

P [A | B] = P [A ∩B]

P [B]
P [B] > 0

Law of Total Probability

P [B] =

n
�

i=1
P [B|Ai] P [Ai]
Ω=

n
�

i=1
Ai

Bayes' Theorem

P [Ai | B] =
P [B | Ai] P [Ai]
�n
j=1 P [B | Aj] P [Aj]
Ω=

n
�

i=1
Ai

Inclusion-Exclusion Principle
����

n
�

i=1
Ai

����=

n
�

r=1
(-1)r-1
�

i≤i1<···<ir≤n

����

r�

j=1
Aij

����

## 3 Random Variables

Random Variable (RV)
X : Ω→R

Probability Mass Function (PMF)

fX(x) = P [X = x] = P [{ω ∈Ω: X(ω) = x}]

Probability Density Function (PDF)

P [a ≤X ≤b] =
�b

a
f(x) dx

Cumulative Distribution Function (CDF)

FX : R →[0, 1]
FX(x) = P [X ≤x]

1. Nondecreasing: x1 < x2 =⇒F(x1) ≤F(x2)
2. Normalized: limx→-∞= 0 and limx→∞= 1
3. Right-Continuous: limy↓x F(y) = F(x)

P [a ≤Y ≤b | X = x] =
�b

a
fY |X(y | x)dy
a ≤b

fY |X(y | x) = f(x, y)

fX(x)

Independence

1. P [X ≤x, Y ≤y] = P [X ≤x] P [Y ≤y]
2. fX,Y (x, y) = fX(x)fY (y)

### 3.1 Transformations

Transformation function
Z = ϕ(X)

Discrete

fZ(z) = P [ϕ(X) = z] = P [{x : ϕ(x) = z}] = P
�
X ∈ϕ-1(z)
�
=
�

x∈ϕ-1(z)
f(x)

Continuous

FZ(z) = P [ϕ(X) ≤z] =
�

Az
f(x) dx
with Az = {x : ϕ(x) ≤z}

Special case if ϕ strictly monotone

fZ(z) = fX(ϕ-1(z))
����
d
dz ϕ-1(z)
����= fX(x)
����
dx
dz

����= fX(x) 1
|J|

The Rule of the Lazy Statistician

E [Z] =
�
ϕ(x) dFX(x)

E [IA(x)] =
�
IA(x) dFX(x) =
�

A
dFX(x) = P [X ∈A]

Convolution

• Z := X + Y
fZ(z) =
�∞

-∞
fX,Y (x, z -x) dx
X,Y ≥0
=
�z

0
fX,Y (x, z -x) dx

• Z := |X -Y |
fZ(z) = 2
�∞

0
fX,Y (x, z + x) dx

• Z := X
Y
fZ(z) =
�∞

-∞
|x|fX,Y (x, xz) dx
⊥⊥
=
�∞

-∞
xfx(x)fX(x)fY (xz) dx

## 4 Expectation

Definition and properties

• E [X] = µX =
�
x dFX(x) =

�

x
xfX(x)
X discrete

�
xfX(x)
X continuous

• P [X = c] = 1 =⇒E [c] = c
• E [cX] = c E [X]
• E [X + Y ] = E [X] + E [Y ]

• E [XY ] =
�

X,Y
xyfX,Y (x, y) dFX(x) dFY (y)

• E [ϕ(Y )] ̸= ϕ(E [X])
(cf. Jensen inequality)
• P [X ≥Y ] = 0 =⇒E [X] ≥E [Y ] ∧P [X = Y ] = 1 =⇒E [X] = E [Y ]

• E [X] =

∞
�

x=1
P [X ≥x]

Sample mean

¯Xn = 1

n

n
�

i=1
Xi

Conditional expectation

• E [Y | X = x] =
�
yf(y | x) dy

• E [X] = E [E [X | Y ]]

• E[ϕ(X, Y ) | X = x] =
�∞

-∞
ϕ(x, y)fY |X(y | x) dx

• E [ϕ(Y, Z) | X = x] =
�∞

-∞
ϕ(y, z)f(Y,Z)|X(y, z | x) dy dz

• E [Y + Z | X] = E [Y | X] + E [Z | X]
• E [ϕ(X)Y | X] = ϕ(X)E [Y | X]
• E[Y | X] = c =⇒Cov [X, Y ] = 0

## 5 Variance

Definition and properties

• V [X] = σ2
X = E
�
(X -E [X])2�
= E
�
X2�
-E [X]2

• V

�n
�

i=1
Xi

�

=

n
�

i=1
V [Xi] + 2
�

i̸=j
Cov [Xi, Yj]

• V

�n
�

i=1
Xi

�

=

n
�

i=1
V [Xi]
iff Xi ⊥⊥Xj

Standard deviation
sd[X] =
�
V [X] = σX

Covariance

• Cov [X, Y ] = E [(X -E [X])(Y -E [Y ])] = E [XY ] -E [X] E [Y ]
• Cov [X, a] = 0
• Cov [X, X] = V [X]
• Cov [X, Y ] = Cov [Y, X]
• Cov [aX, bY ] = abCov [X, Y ]

• Cov [X + a, Y + b] = Cov [X, Y ]

• Cov

n
�

i=1
Xi,

m
�

j=1
Yj

=

n
�

i=1

m
�

j=1
Cov [Xi, Yj]

Correlation

ρ [X, Y ] =
Cov [X, Y ]
�
V [X] V [Y ]

Independence

X ⊥⊥Y =⇒ρ [X, Y ] = 0 ⇐⇒Cov [X, Y ] = 0 ⇐⇒E [XY ] = E [X] E [Y ]

Sample variance

S2 =
1
n -1

n
�

i=1
(Xi -¯Xn)2

Conditional variance

• V [Y | X] = E
�
(Y -E [Y | X])2 | X
�
= E
�
Y 2 | X
�
-E [Y | X]2

• V [Y ] = E [V [Y | X]] + V [E [Y | X]]

## 6 Inequalities

Cauchy-Schwarz
E [XY ]2 ≤E
�
X2�
E
�
Y 2�

Markov

P [ϕ(X) ≥t] ≤E [ϕ(X)]

t
Chebyshev

P [|X -E [X]| ≥t] ≤V [X]

t2

Chernoff

P [X ≥(1 + δ)µ] ≤
�
eδ

(1 + δ)1+δ

�
δ > -1

Jensen
E [ϕ(X)] ≥ϕ(E [X])
ϕ convex

## 7 Distribution Relationships

Binomial

• Xi ∼Bern (p) =⇒

n
�

i=1
Xi ∼Bin (n, p)

• X ∼Bin (n, p) , Y ∼Bin (m, p) =⇒X + Y ∼Bin (n + m, p)

• limn→∞Bin (n, p) = Po (np)
(n large, p small)
• limn→∞Bin (n, p) = N (np, np(1 -p))
(n large, p far from 0 and 1)

Negative Binomial

• X ∼NBin (1, p) = Geo (p)
• X ∼NBin (r, p) = �r
i=1 Geo (p)
• Xi ∼NBin (ri, p) =⇒�Xi ∼NBin (�ri, p)
• X ∼NBin (r, p) . Y ∼Bin (s + r, p) =⇒P [X ≤s] = P [Y ≥r]

Poisson

• Xi ∼Po (λi) ∧Xi ⊥⊥Xj =⇒

n
�

i=1
Xi ∼Po

�n
�

i=1
λi

�

• Xi ∼Po (λi) ∧Xi ⊥⊥Xj =⇒Xi

������

n
�

j=1
Xj ∼Bin

n
�

j=1
Xj,
λi
�n
j=1 λj

Exponential

• Xi ∼Exp (β) ∧Xi ⊥⊥Xj =⇒

n
�

i=1
Xi ∼Gamma (n, β)

• Memoryless property: P [X > x + y | X > y] = P [X > x]

Normal

• X ∼N
�
µ, σ2�
=⇒
�
X-µ

σ
�
∼N (0, 1)

• X ∼N
�
µ, σ2�
∧Z = aX + b =⇒Z ∼N
�
aµ + b, a2σ2�

• X ∼N
�
µ1, σ2
1
�
∧Y ∼N
�
µ2, σ2
2
�
=⇒X + Y ∼N
�
µ1 + µ2, σ2
1 + σ2
2
�

• Xi ∼N
�
µi, σ2
i
�
=⇒�

i Xi ∼N
��

i µi, �

i σ2
i
�

• P [a < X ≤b] = Φ
�
b-µ

σ
�
-Φ
�a-µ
σ
�

• Φ(-x) = 1 -Φ(x)
φ′(x) = -xφ(x)
φ′′(x) = (x2 -1)φ(x)
• Upper quantile of N (0, 1): zα = Φ-1(1 -α)

Gamma

• X ∼Gamma (α, β) ⇐⇒X/β ∼Gamma (α, 1)
• Gamma (α, β) ∼�α
i=1 Exp (β)
• Xi ∼Gamma (αi, β) ∧Xi ⊥⊥Xj =⇒�

i Xi ∼Gamma (�

i αi, β)

• Γ(α)
λα
=
�∞

0
xα-1e-λx dx

Beta

•
1
B(α, β)xα-1(1 -x)β-1 = Γ(α + β)

Γ(α)Γ(β)xα-1(1 -x)β-1

• E
�
Xk�
= B(α + k, β)

B(α, β)
=
α + k -1
α + β + k -1E
�
Xk-1�

• Beta (1, 1) ∼Unif (0, 1)

## 8 Probability and Moment Generating Functions

• GX(t) = E
�
tX�
|t| < 1

• MX(t) = GX(et) = E
�
eXt�
= E

�∞
�

i=0

(Xt)i

i!

�

=

∞
�

i=0

E
�
Xi�

i!
· ti

• P [X = 0] = GX(0)
• P [X = 1] = G′
X(0)

• P [X = i] = G(i)
X (0)

i!
• E [X] = G′
X(1-)

• E
�
Xk�
= M (k)
X (0)

• E
�
X!
(X -k)!

�
= G(k)
X (1-)

• V [X] = G′′
X(1-) + G′
X(1-) -(G′
X(1-))2

• GX(t) = GY (t) =⇒X
d= Y

## 9 Multivariate Distributions

### 9.1 Standard Bivariate Normal

Let X, Y ∼N (0, 1) ∧X ⊥⊥Z where Y = ρX +
�
1 -ρ2Z

Joint density

f(x, y) =
1

2π
�
1 -ρ2 exp
�
-x2 + y2 -2ρxy

2(1 -ρ2)

�

Conditionals

(Y | X = x) ∼N
�
ρx, 1 -ρ2�
and
(X | Y = y) ∼N
�
ρy, 1 -ρ2�

Independence
X ⊥⊥Y ⇐⇒ρ = 0

### 9.2 Bivariate Normal

Let X ∼N
�
µx, σ2
x
�
and Y ∼N
�
µy, σ2
y
�
.

f(x, y) =
1

2πσxσy
�
1 -ρ2 exp
�
z
2(1 -ρ2)

�

z =

��x -µx
σx

�2
+
�y -µy
σy

�2
-2ρ
�x -µx
σx

��y -µy
σy

��

Conditional mean and variance

E [X | Y ] = E [X] + ρσX

σY
(Y -E [Y ])

V [X | Y ] = σX
�
1 -ρ2

### 9.3 Multivariate Normal

Covariance matrix Σ
(Precision matrix Σ-1)

Σ =

V [X1]
· · ·
Cov [X1, Xk]
...
...
...
Cov [Xk, X1]
· · ·
V [Xk]

If X ∼N (µ, Σ),

fX(x) = (2π)-n/2 |Σ|-1/2 exp
�
-1

2(x -µ)T Σ-1(x -µ)
�

Properties

• Z ∼N (0, 1) ∧X = µ + Σ1/2Z =⇒X ∼N (µ, Σ)
• X ∼N (µ, Σ) =⇒Σ-1/2(X -µ) ∼N (0, 1)
• X ∼N (µ, Σ) =⇒AX ∼N
�
Aµ, AΣAT �

• X ∼N (µ, Σ) ∧∥a∥= k =⇒aT X ∼N
�
aT µ, aT Σa
�

## 10 Convergence

Let {X1, X2, . . .} be a sequence of rv's and let X be another rv. Let Fn denote
the cdf of Xn and let F denote the cdf of X.

Types of convergence

1. In distribution (weakly, in law): Xn
D→X

lim
n→∞Fn(t) = F(t)
∀t where F continuous

2. In probability: Xn
P→X

(∀ε > 0) lim
n→∞P [|Xn -X| > ε] = 0

3. Almost surely (strongly): Xn
as
→X

P
�
lim
n→∞Xn = X
�
= P
�
ω ∈Ω: lim
n→∞Xn(ω) = X(ω)
�
= 1

4. In quadratic mean (L2): Xn
qm
→X

lim
n→∞E
�
(Xn -X)2�
= 0

Relationships

• Xn
qm
→X =⇒Xn
P→X =⇒Xn
D→X
• Xn
as
→X =⇒Xn
P→X
• Xn
D→X ∧(∃c ∈R) P [X = c] = 1 =⇒Xn
P→X

• Xn
P→X ∧Yn
P→Y =⇒Xn + Yn
P→X + Y
• Xn
qm
→X ∧Yn
qm
→Y =⇒Xn + Yn
qm
→X + Y
• Xn
P→X ∧Yn
P→Y =⇒XnYn
P→XY
• Xn
P→X =⇒ϕ(Xn)
P→ϕ(X)

• Xn
D→X =⇒ϕ(Xn)
D→ϕ(X)
• Xn
qm
→b ⇐⇒limn→∞E [Xn] = b ∧limn→∞V [Xn] = 0
• X1, . . . , Xn iid ∧E [X] = µ ∧V [X] < ∞⇐⇒
¯Xn
qm
→µ

Slutzky's Theorem

• Xn
D→X and Yn
P→c =⇒Xn + Yn
D→X + c
• Xn
D→X and Yn
P→c =⇒XnYn
D→cX
• In general: Xn
D→X and Yn
D→Y ̸=⇒Xn + Yn
D→X + Y

### 10.1 Law of Large Numbers (LLN)

Let {X1, . . . , Xn} be a sequence of iid rv's, E [X1] = µ, and V [X1] < ∞.

Weak (WLLN)
¯Xn
P→µ
n →∞

Strong (WLLN)
¯Xn
as
→µ
n →∞

### 10.2 Central Limit Theorem (CLT)

Let {X1, . . . , Xn} be a sequence of iid rv's, E [X1] = µ, and V [X1] = σ2.

Zn :=
¯Xn -µ
�
V
�¯Xn
�=
√n( ¯Xn -µ)

σ

D→Z
where Z ∼N (0, 1)

lim
n→∞P [Zn ≤z] = Φ(z)
z ∈R

CLT notations

Zn ≈N (0, 1)

¯Xn ≈N
�
µ, σ2

n

�

¯Xn -µ ≈N
�
0, σ2

n

�

√n( ¯Xn -µ) ≈N
�
0, σ2�

√n( ¯Xn -µ)

n
≈N (0, 1)

Continuity correction

P
�¯Xn ≤x
�
≈Φ
�x + 1
2 -µ
σ/√n

�

P
�¯Xn ≥x
�
≈1 -Φ
�x -1
2 -µ
σ/√n

�

Delta method

Yn ≈N
�
µ, σ2

n

�
=⇒ϕ(Yn) ≈N
�
ϕ(µ), (ϕ′(µ))2 σ2

n

�

## 11 Statistical Inference

Let X1, · · · , Xn
iid
∼F if not otherwise noted.

### 11.1 Point Estimation

• Point estimator �θn of θ is a rv: �θn = g(X1, . . . , Xn)

• bias(�θn) = E
�
�θn
�
-θ

• Consistency: �θn
P→θ
• Sampling distribution: F(�θn)

• Standard error: se(�θn) =
�
V
�
�θn
�

• Mean squared error: mse = E
�
(�θn -θ)2�
= bias(�θn)2 + V
�
�θn
�

• limn→∞bias(�θn) = 0 ∧limn→∞se(�θn) = 0 =⇒�θn is consistent

• Asymptotic normality:
�θn -θ
se

D→N (0, 1)

• Slutzky's Theorem often lets us replace se(�θn) by some (weakly) consistent estimator �σn.

### 11.2 Normal-Based Confidence Interval

Suppose �θn ≈N
�
θ, �se2�
. Let zα/2 = Φ-1(1 -(α/2)), i.e., P
�
Z > zα/2
�
= α/2

and P
�
-zα/2 < Z < zα/2
�
= 1 -α where Z ∼N (0, 1). Then

Cn = �θn ± zα/2 �se

### 11.3 Empirical distribution

Empirical Distribution Function (ECDF)

�Fn(x) =
�n
i=1 I(Xi ≤x)

n

I(Xi ≤x) =

�
1
Xi ≤x
0
Xi > x

Properties (for any fixed x)

• E
�
�Fn
�
= F(x)

• V
�
�Fn
�
= F(x)(1 -F(x))

n

• mse = F(x)(1 -F(x))
n

D→0

• �Fn
P→F(x)

Dvoretzky-Kiefer-Wolfowitz (DKW) inequality (X1, . . . , Xn ∼F)

P
�
sup
x

���F(x) -�Fn(x)
���> ε
�
= 2e-2nε2

Nonparametric 1 -α confidence band for F

L(x) = max{ �Fn -ϵn, 0}

U(x) = min{ �Fn + ϵn, 1}

ϵ =

�
1
2n log
�2
α

�

P [L(x) ≤F(x) ≤U(x) ∀x] ≥1 -α

### 11.4 Statistical Functionals

• Statistical functional: T(F)
• Plug-in estimator of θ = (F): �θn = T( �Fn)
• Linear functional: T(F) =
�
ϕ(x) dFX(x)
• Plug-in estimator for linear functional:

T( �Fn) =
�
ϕ(x) d �Fn(x) = 1

n

n
�

i=1
ϕ(Xi)

• Often: T( �Fn) ≈N
�
T(F), �se2�
=⇒T( �Fn) ± zα/2 �se

• pth quantile: F -1(p) = inf{x : F(x) ≥p}
• �µ = ¯Xn

• �σ2 =
1
n -1

n
�

i=1
(Xi -¯Xn)2

• �κ =

1
n
�n
i=1(Xi -�µ)3

�σ3j

• �ρ =
�n
i=1(Xi -¯Xn)(Yi -¯Yn)
��n
i=1(Xi -¯Xn)2
��n
i=1(Yi -¯Yn)

## 12 Parametric Inference

Let F =
�
f(x; θ) : θ ∈Θ
�
be a parametric model with parameter space Θ ⊂Rk

and parameter θ = (θ1, . . . , θk).

### 12.1 Method of Moments

jth moment

αj(θ) = E
�
Xj�
=
�
xj dFX(x)

jth sample moment

�αj = 1
n

n
�

i=1
Xj
i

Method of moments estimator (MoM)

α1(θ) = �α1
α2(θ) = �α2
... = ...

αk(θ) = �αk

Properties of the MoM estimator

• �θn exists with probability tending to 1
• Consistency: �θn
P→θ
• Asymptotic normality:

√n(�θ -θ)
D→N (0, Σ)

where Σ = gE
�
Y Y T �
gT , Y = (X, X2, . . . , Xk)T ,
g = (g1, . . . , gk) and gj =
∂
∂θα-1
j (θ)

### 12.2 Maximum Likelihood

Likelihood: Ln : Θ →[0, ∞)

Ln(θ) =

n
�

i=1
f(Xi; θ)

Log-likelihood

ℓn(θ) = log Ln(θ) =

n
�

i=1
log f(Xi; θ)

Maximum likelihood estimator (mle)

Ln(�θn) = sup
θ
Ln(θ)

Score function
s(X; θ) = ∂

∂θ log f(X; θ)

Fisher information
I(θ) = Vθ [s(X; θ)]

In(θ) = nI(θ)

Fisher information (exponential family)

I(θ) = Eθ

�
-∂

∂θs(X; θ)
�

Observed Fisher information

Iobs
n (θ) = -∂2

∂θ2

n
�

i=1
log f(Xi; θ)

Properties of the mle

• Consistency: �θn
P→θ

• Equivariance: �θn is the mle =⇒ϕ(�θn) ist the mle of ϕ(θ)
• Asymptotic normality:

1. se ≈
�
1/In(θ)
(�θn -θ)

se

D→N (0, 1)

2. �se ≈
�
1/In(�θn)

(�θn -θ)

�se

D→N (0, 1)

• Asymptotic optimality (or efficiency), i.e., smallest variance for large samples. If �θn is any other estimator, the asymptotic relative efficiency is

are(�θn, �θn) =
V
�
�θn
�

V
�
�θn
�≤1

• Approximately the Bayes estimator

#### 12.2.1 Delta Method

If τ = ϕ(�θ) where ϕ is differentiable and ϕ′(θ) ̸= 0:

(�τn -τ)

�se(�τ)

D→N (0, 1)

where �τ = ϕ(�θ) is the mle of τ and

�se =
���ϕ′(�θ)
����se(�θn)

### 12.3 Multiparameter Models

Let θ = (θ1, . . . , θk) and �θ = (�θ1, . . . , �θk) be the mle.

Hjj = ∂2ℓn

∂θ2
Hjk =
∂2ℓn
∂θj∂θk

Fisher information matrix

In(θ) = -

Eθ [H11]
· · ·
Eθ [H1k]
...
...
...
Eθ [Hk1]
· · ·
Eθ [Hkk]

Under appropriate regularity conditions

(�θ -θ) ≈N (0, Jn)

with Jn(θ) = I-1
n . Further, if �θj is the jth component of θ, then

(�θj -θj)

�sej

D→N (0, 1)

where �se2
j = Jn(j, j) and Cov
�
�θj, �θk
�
= Jn(j, k)

#### 12.3.1 Multiparameter delta method

Let τ = ϕ(θ1, . . . , θk) and let the gradient of ϕ be

∇ϕ =

∂ϕ
∂θ1
...
∂ϕ
∂θk

Suppose ∇ϕ
��
θ=�θ ̸= 0 and �τ = ϕ(�θ). Then,

(�τ -τ)

�se(�τ)

D→N (0, 1)

where

�se(�τ) =

��
�∇ϕ
�T �Jn
�
�∇ϕ
�

and �Jn = Jn(�θ) and �∇ϕ = ∇ϕ
��
θ=�θ.

### 12.4 Parametric Bootstrap

Sample from f(x; �θn) instead of from �Fn, where �θn could be the mle or method
of moments estimator.

## 13 Hypothesis Testing

H0 : θ ∈Θ0
versus
H1 : θ ∈Θ1

Definitions

• Null hypothesis H0
• Alternative hypothesis H1
• Simple hypothesis θ = θ0
• Composite hypothesis θ > θ0 or θ < θ0
• Two-sided test: H0 : θ = θ0
versus
H1 : θ ̸= θ0
• One-sided test: H0 : θ ≤θ0
versus
H1 : θ > θ0
• Critical value c
• Test statistic T
• Rejection region R = {x : T(x) > c}
• Power function β(θ) = P [X ∈R]
• Power of a test: 1 -P [Type II error] = 1 -β = inf
θ∈Θ1 β(θ)

• Test size: α = P [Type I error] = sup
θ∈Θ0
β(θ)

Retain H0
Reject H0
H0 true
√
Type I Error (α)
H1 true
Type II Error (β)
√(power)
p-value

• p-value = supθ∈Θ0 Pθ [T(X) ≥T(x)] = inf
�
α : T(x) ∈Rα
�

• p-value = supθ∈Θ0
Pθ [T(X⋆) ≥T(X)]
�
��
�
1-Fθ(T (X))
since T (X⋆)∼Fθ

= inf
�
α : T(X) ∈Rα
�

p-value
evidence
< 0.01
very strong evidence against H0
0.01 -0.05
strong evidence against H0
0.05 -0.1
weak evidence against H0
> 0.1
little or no evidence against H0
Wald test

• Two-sided test

• Reject H0 when |W| > zα/2 where W =
�θ -θ0
�se
• P
�
|W| > zα/2
�
→α
• p-value = Pθ0 [|W| > |w|] ≈P [|Z| > |w|] = 2Φ(-|w|)

Likelihood ratio test (LRT)

• T(X) = supθ∈Θ Ln(θ)
supθ∈Θ0 Ln(θ) = Ln(�θn)

Ln(�θn,0)

• λ(X) = 2 log T(X)
D→χ2
r-q where

k
�

i=1
Z2
i ∼χ2
k and Z1, . . . , Zk
iid
∼N (0, 1)

• p-value = Pθ0 [λ(X) > λ(x)] ≈P
�
χ2
r-q > λ(x)
�

Multinomial LRT

• mle: �pn =
�X1
n , . . . , Xk

n

�

• T(X) = Ln(�pn)
Ln(p0) =

k
�

j=1

��pj
p0j

�Xj

• λ(X) = 2

k
�

j=1
Xj log
��pj
p0j

�

D→χ2
k-1

• The approximate size α LRT rejects H0 when λ(X) ≥χ2
k-1,α
Pearson Chi-square Test

• T =

k
�

j=1

(Xj -E [Xj])2

E [Xj]
where E [Xj] = np0j under H0

• T
D→χ2
k-1
• p-value = P
�
χ2
k-1 > T(x)
�

• Faster
D→X2
k-1 than LRT, hence preferable for small n

Independence testing

• I rows, J columns, X multinomial sample of size n = I ∗J
• mles unconstrained: �pij = Xij
n
• mles under H0: �p0ij = �pi·�p·j = Xi·
n
X·j

n
• LRT: λ = 2 �I
i=1
�J
j=1 Xij log
�
nXij
Xi·X·j

�

• PearsonChiSq: T = �I
i=1
�J
j=1
(Xij-E[Xij])2

E[Xij]
• LRT and Pearson
D→χ2
kν, where ν = (I -1)(J -1)

## 14 Bayesian Inference

Bayes' Theorem

f(θ | x) = f(x | θ)f(θ)

f(xn)
=
f(x | θ)f(θ)
�
f(x | θ)f(θ) dθ ∝Ln(θ)f(θ)

Definitions

• Xn = (X1, . . . , Xn)

• xn = (x1, . . . , xn)
• Prior density f(θ)
• Likelihood f(xn | θ): joint density of the data

In particular, Xn iid =⇒f(xn | θ) =

n
�

i=1
f(xi | θ) = Ln(θ)

• Posterior density f(θ | xn)
• Normalizing constant cn = f(xn) =
�
f(x | θ)f(θ) dθ
• Kernel: part of a density that depends on θ

• Posterior mean ¯θn =
�
θf(θ | xn) dθ =

�
θLn(θ)f(θ)
�
Ln(θ)f(θ) dθ

### 14.1 Credible Intervals

Posterior interval

P [θ ∈(a, b) | xn] =
�b

a
f(θ | xn) dθ = 1 -α

Equal-tail credible interval

�a

-∞
f(θ | xn) dθ =
�∞

b
f(θ | xn) dθ = α/2

Highest posterior density (HPD) region Rn

1. P [θ ∈Rn] = 1 -α
2. Rn = {θ : f(θ | xn) > k} for some k

Rn is unimodal =⇒Rn is an interval

### 14.2 Function of parameters

Let τ = ϕ(θ) and A = {θ : ϕ(θ) ≤τ}.
Posterior CDF for τ

H(r | xn) = P [ϕ(θ) ≤τ | xn] =
�

A
f(θ | xn) dθ

Posterior density
h(τ | xn) = H′(τ | xn)

Bayesian delta method

τ | Xn ≈N
�
ϕ(�θ), �se
���ϕ′(�θ)
���
�

### 14.3 Priors

Choice

• Subjective bayesianism.
• Objective bayesianism.
• Robust bayesianism.

Types

• Flat: f(θ) ∝constant
• Proper:
�∞
-∞f(θ) dθ = 1

• Improper:
�∞
-∞f(θ) dθ = ∞
• Jeffrey's prior (transformation-invariant):

f(θ) ∝
�
I(θ)
f(θ) ∝
�
det(I(θ))

• Conjugate: f(θ) and f(θ | xn) belong to the same parametric family

#### 14.3.1 Conjugate Priors

Discrete likelihood

Likelihood
Conjugate prior
Posterior hyperparameters

Bern (p)
Beta (α, β)
α +

n
�

i=1
xi, β + n -

n
�

i=1
xi

Bin (p)
Beta (α, β)
α +

n
�

i=1
xi, β +

n
�

i=1
Ni -

n
�

i=1
xi

NBin (p)
Beta (α, β)
α + rn, β +

n
�

i=1
xi

Po (λ)
Gamma (α, β)
α +

n
�

i=1
xi, β + n

Multinomial(p)
Dir (α)
α +

n
�

i=1
x(i)

Geo (p)
Beta (α, β)
α + n, β +

n
�

i=1
xi

Continuous likelihood (subscript c denotes constant)

Likelihood
Conjugate prior
Posterior hyperparameters

Unif (0, θ)
Pareto(xm, k)
max
�
x(n), xm
�
, k + n

Exp (λ)
Gamma (α, β)
α + n, β +

n
�

i=1
xi

N
�
µ, σ2
c
�
N
�
µ0, σ2
0
�
�µ0
σ2
0
+
�n
i=1 xi

σ2c

�
/
�1
σ2
0
+ n

σ2c

�
,
�1
σ2
0
+ n

σ2c

�-1

N
�
µc, σ2�

Scaled Inverse Chisquare(ν, σ2
0)

ν + n, νσ2
0 + �n
i=1(xi -µ)2

ν + n

N
�
µ, σ2�

Normalscaled
Inverse
Gamma(λ, ν, α, β)

νλ + n¯x

ν + n ,
ν + n,
α + n

2 ,

β + 1

2

n
�

i=1
(xi -¯x)2 + γ(¯x -λ)2

2(n + γ)

MVN(µ, Σc)
MVN(µ0, Σ0)

�
Σ-1
0
+ nΣ-1
c
�-1 �
Σ-1
0 µ0 + nΣ-1¯x
�
,
�
Σ-1
0
+ nΣ-1
c
�-1

MVN(µc, Σ)
Inverse-
Wishart(κ, Ψ)

n + κ, Ψ +

n
�

i=1
(xi -µc)(xi -µc)T

Pareto(xmc, k)
Gamma (α, β)
α + n, β +

n
�

i=1
log xi

xmc
Pareto(xm, kc)
Pareto(x0, k0)
x0, k0 -kn where k0 > kn

Gamma (αc, β)
Gamma (α0, β0)
α0 + nαc, β0 +

n
�

i=1
xi

### 14.4 Bayesian Testing

If H0 : θ ∈Θ0:

Prior probability P [H0] =
�

Θ0
f(θ) dθ

Posterior probability P [H0 | xn] =
�

Θ0
f(θ | xn) dθ

Let H0, . . . , HK-1 be K hypotheses. Suppose θ ∼f(θ | Hk),

P [Hk | xn] =
f(xn | Hk)P [Hk]
�K
k=1 f(xn | Hk)P [Hk]
,

Marginal likelihood

f(xn | Hi) =
�

Θ
f(xn | θ, Hi)f(θ | Hi) dθ

Posterior odds (of Hi relative to Hj)

P [Hi | xn]
P [Hj | xn]
=
f(xn | Hi)
f(xn | Hj)
�
��
�
Bayes Factor BFij

×
P [Hi]
P [Hj]
����
prior odds

Bayes factor
log10 BF10
BF10
evidence

0 -0.5
1 -1.5
Weak
0.5 -1
1.5 -10
Moderate
1 -2
10 -100
Strong
> 2
> 100
Decisive

p∗=

p
1-pBF10
1 +
p
1-pBF10
where p = P [H1] and p∗= P [H1 | xn]

## 15 Exponential Family

Scalar parameter

fX(x | θ) = h(x) exp {η(θ)T(x) -A(θ)}

= h(x)g(θ) exp {η(θ)T(x)}

Vector parameter

fX(x | θ) = h(x) exp

�s
�

i=1
ηi(θ)Ti(x) -A(θ)

�

= h(x) exp {η(θ) · T(x) -A(θ)}

= h(x)g(θ) exp {η(θ) · T(x)}

Natural form

fX(x | η) = h(x) exp {η · T(x) -A(η)}

= h(x)g(η) exp {η · T(x)}

= h(x)g(η) exp
�
ηT T(x)
�

## 16 Sampling Methods

### 16.1 The Bootstrap

Let Tn = g(X1, . . . , Xn) be a statistic.

1. Estimate VF [Tn] with V �
Fn [Tn].
2. Approximate V �
Fn [Tn] using simulation:

(a) Repeat the following B times to get T ∗
n,1, . . . , T ∗
n,B, an iid sample from
the sampling distribution implied by �Fn

i. Sample uniformly X∗
1, . . . , X∗
n ∼�Fn.
ii. Compute T ∗
n = g(X∗
1, . . . , X∗
n).

(b) Then

vboot = �V �
Fn = 1

B

B
�

b=1

�

T ∗
n,b -1

B

B
�

r=1
T ∗
n,r

�2

#### 16.1.1 Bootstrap Confidence Intervals

Normal-based interval

Tn ± zα/2 �seboot

Pivotal interval

1. Location parameter θ = T(F)
2. Pivot Rn = �θn -θ
3. Let H(r) = P [Rn ≤r] be the cdf of Rn
4. Let R∗
n,b = �θ∗
n,b -�θn. Approximate H using bootstrap:

�H(r) = 1
B

B
�

b=1
I(R∗
n,b ≤r)

5. θ∗
β = β sample quantile of (�θ∗
n,1, . . . , �θ∗
n,B)

6. r∗
β = β sample quantile of (R∗
n,1, . . . , R∗
n,B), i.e., r∗
β = θ∗
β -�θn

7. Approximate 1 -α confidence interval Cn =
�
ˆa,ˆb
�
where

ˆa =
�θn -�H-1 �
1 -α

2

�
=
�θn -r∗
1-α/2 =
2�θn -θ∗
1-α/2

ˆb =
�θn -�H-1 �α
2

�
=
�θn -r∗
α/2 =
2�θn -θ∗
α/2

Percentile interval

Cn =
�
θ∗
α/2, θ∗
1-α/2
�

### 16.2 Rejection Sampling

Setup

• We can easily sample from g(θ)
• We want to sample from h(θ), but it is difficult

• We know h(θ) up to a proportional constant: h(θ) =
k(θ)
�
k(θ) dθ
• Envelope condition: we can find M > 0 such that k(θ) ≤Mg(θ)
∀θ

Algorithm

1. Draw θcand ∼g(θ)
2. Generate u ∼Unif (0, 1)

3. Accept θcand if u ≤
k(θcand)
Mg(θcand)
4. Repeat until B values of θcand have been accepted

Example

• We can easily sample from the prior g(θ) = f(θ)
• Target is the posterior h(θ) ∝k(θ) = f(xn | θ)f(θ)
• Envelope condition: f(xn | θ) ≤f(xn | �θn) = Ln(�θn) ≡M
• Algorithm

1. Draw θcand ∼f(θ)
2. Generate u ∼Unif (0, 1)

3. Accept θcand if u ≤Ln(θcand)

Ln(�θn)

### 16.3 Importance Sampling

Sample from an importance function g rather than target density h.
Algorithm to obtain an approximation to E [q(θ) | xn]:

1. Sample from the prior θ1, . . . , θn
iid
∼f(θ)

2. wi =
Ln(θi)
�B
i=1 Ln(θi)
∀i = 1, . . . , B

3. E [q(θ) | xn] ≈�B
i=1 q(θi)wi

## 17 Decision Theory

Definitions

• Unknown quantity affecting our decision: θ ∈Θ

• Decision rule: synonymous for an estimator �θ
• Action a ∈A: possible value of the decision rule. In the estimation
context, the action is just an estimate of θ, �θ(x).
• Loss function L: consequences of taking action a when true state is θ or
discrepancy between θ and �θ, L : Θ × A →[-k, ∞).

Loss functions

• Squared error loss: L(θ, a) = (θ -a)2

• Linear loss: L(θ, a) =

�
K1(θ -a)
a -θ < 0
K2(a -θ)
a -θ ≥0
• Absolute error loss: L(θ, a) = |θ -a|
(linear loss with K1 = K2)
• Lp loss: L(θ, a) = |θ -a|p

• Zero-one loss: L(θ, a) =

�
0
a = θ
1
a ̸= θ

### 17.1 Risk

Posterior risk

r(�θ | x) =
�
L(θ, �θ(x))f(θ | x) dθ = Eθ|X
�
L(θ, �θ(x))
�

(Frequentist) risk

R(θ, �θ) =
�
L(θ, �θ(x))f(x | θ) dx = EX|θ
�
L(θ, �θ(X))
�

Bayes risk

r(f, �θ) =
��
L(θ, �θ(x))f(x, θ) dx dθ = Eθ,X
�
L(θ, �θ(X))
�

r(f, �θ) = Eθ
�
EX|θ
�
L(θ, �θ(X)
��
= Eθ
�
R(θ, �θ)
�

r(f, �θ) = EX
�
Eθ|X
�
L(θ, �θ(X)
��
= EX
�
r(�θ | X)
�

### 17.2 Admissibility

• �θ′ dominates �θ if
∀θ : R(θ, �θ′) ≤R(θ, �θ)

∃θ : R(θ, �θ′) < R(θ, �θ)

• �θ is inadmissible if there is at least one other estimator �θ′ that dominates
it. Otherwise it is called admissible.

### 17.3 Bayes Rule

Bayes rule (or Bayes estimator)

• r(f, �θ) = inf �θ r(f, �θ)

• �θ(x) = inf r(�θ | x) ∀x =⇒r(f, �θ) =
�
r(�θ | x)f(x) dx

Theorems

• Squared error loss: posterior mean
• Absolute error loss: posterior median
• Zero-one loss: posterior mode

### 17.4 Minimax Rules

Maximum risk
¯R(�θ) = sup
θ
R(θ, �θ)
¯R(a) = sup
θ
R(θ, a)

Minimax rule
sup
θ
R(θ, �θ) = inf
�θ
¯R(�θ) = inf
�θ
sup
θ
R(θ, �θ)

�θ = Bayes rule ∧∃c : R(θ, �θ) = c

Least favorable prior

�θf = Bayes rule ∧R(θ, �θf) ≤r(f, �θf) ∀θ

## 18 Linear Regression

Definitions

• Response variable Y
• Covariate X (aka predictor variable or feature)

### 18.1 Simple Linear Regression

Model
Yi = β0 + β1Xi + ϵi
E [ϵi | Xi] = 0, V [ϵi | Xi] = σ2

Fitted line
�r(x) = �β0 + �β1x

Predicted (fitted) values
�Yi = �r(Xi)

Residuals
ˆϵi = Yi -�Yi = Yi -
�
�β0 + �β1Xi
�

Residual sums of squares (rss)

rss(�β0, �β1) =

n
�

i=1
ˆϵ2
i

Least square estimates
�βT = (�β0, �β1)T : min
�β0,�β1
rss

�β0 = ¯Yn -�β1 ¯Xn

�β1 =
�n
i=1(Xi -¯Xn)(Yi -¯Yn)
�n
i=1(Xi -¯Xn)2
=
�n
i=1 XiYi -n ¯XY
�n
i=1 X2
i -nX2

E
�
�β | Xn�
=
�β0
β1

�

V
�
�β | Xn�
= σ2

nsX

�
n-1 �n
i=1 X2
i
-Xn
-Xn
1

�

�se(�β0) =
�σ
sX
√n

��n
i=1 X2
i
n

�se(�β1) =
�σ
sX
√n

where s2
X = n-1 �n
i=1(Xi -Xn)2 and �σ2 =
1
n-2
�n
i=1 ˆϵ2
i (unbiased estimate).
Further properties:

• Consistency: �β0
P→β0 and �β1
P→β1
• Asymptotic normality:

�β0 -β0
�se(�β0)

D→N (0, 1)
and
�β1 -β1
�se(�β1)

D→N (0, 1)

• Approximate 1 -α confidence intervals for β0 and β1:

�β0 ± zα/2 �se(�β0)
and
�β1 ± zα/2 �se(�β1)

• Wald test for H0 : β1 = 0 vs. H1 : β1 ̸= 0: reject H0 if |W| > zα/2 where
W = �β1/�se(�β1).

R2

R2 =
�n
i=1(�Yi -Y )2
�n
i=1(Yi -Y )2 = 1 -
�n
i=1 ˆϵ2
i
�n
i=1(Yi -Y )2 = 1 -rss

tss

Likelihood

L =

n
�

i=1
f(Xi, Yi) =

n
�

i=1
fX(Xi) ×

n
�

i=1
fY |X(Yi | Xi) = L1 × L2

L1 =

n
�

i=1
fX(Xi)

L2 =

n
�

i=1
fY |X(Yi | Xi) ∝σ-n exp

�

-1

2σ2
�

i

�
Yi -(β0 -β1Xi)
�2
�

Under the assumption of Normality, the least squares estimator is also the mle

�σ2 = 1
n

n
�

i=1
ˆϵ2
i

### 18.2 Prediction

Observe X = x∗of the covariate and want to predict their outcome Y∗.

�Y∗= �β0 + �β1x∗

V
�
�Y∗
�
= V
�
�β0
�
+ x2
∗V
�
�β1
�
+ 2x∗Cov
�
�β0, �β1
�

Prediction interval

�ξ2
n = �σ2
��n
i=1(Xi -X∗)2

n �
i(Xi -¯X)2j + 1
�

�Y∗± zα/2�ξn

### 18.3 Multiple Regression

Y = Xβ + ϵ

where

X =

X11
· · ·
X1k
...
...
...
Xn1
· · ·
Xnk

β =

β1
...
βk

ϵ =

ϵ1
...
ϵn

Likelihood

L(µ, Σ) = (2πσ2)-n/2 exp
�
-1

2σ2 rss
�

rss = (y -Xβ)T (y -Xβ) = ∥Y -Xβ∥2 =

N
�

i=1
(Yi -xT
i β)2

If the (k × k) matrix XT X is invertible,

�β = (XT X)-1XT Y

V
�
�β | Xn�
= σ2(XT X)-1

�β ≈N
�
β, σ2(XT X)-1�

Estimate regression function

�r(x) =

k
�

j=1
�βjxj

Unbiased estimate for σ2

�σ2 =
1
n -k

n
�

i=1
ˆϵ2
i
ˆϵ = X �β -Y

mle
�µ = ¯X
�σ2 = n -k
n
σ2

1 -α Confidence interval
�βj ± zα/2 �se(�βj)

### 18.4 Model Selection

Consider predicting a new observation Y ∗for covariates X∗and let S ⊂J
denote a subset of the covariates in the model, where |S| = k and |J| = n.
Issues

• Underfitting: too few covariates yields high bias
• Overfitting: too many covariates yields high variance

Procedure

1. Assign a score to each model
2. Search through all models to find the one with the highest score

Hypothesis testing

H0 : βj = 0 vs. H1 : βj ̸= 0
∀j ∈J

Mean squared prediction error (mspe)

mspe = E
�
(�Y (S) -Y ∗)2�

Prediction risk

R(S) =

n
�

i=1
mspei =

n
�

i=1
E
�
(�Yi(S) -Y ∗
i )2�

Training error

�Rtr(S) =

n
�

i=1
(�Yi(S) -Yi)2

R2

R2(S) = 1 -rss(S)

tss
= 1 -
�Rtr(S)
tss
= 1 -
�n
i=1(�Yi(S) -Y )2
�n
i=1(Yi -Y )2

The training error is a downward-biased estimate of the prediction risk.

E
�
�Rtr(S)
�
< R(S)

bias( �Rtr(S)) = E
�
�Rtr(S)
�
-R(S) = -2

n
�

i=1
Cov
�
�Yi, Yi
�

Adjusted R2

R2(S) = 1 -n -1

n -k
rss
tss
Mallow's Cp statistic

�R(S) = �Rtr(S) + 2k�σ2 = lack of fit + complexity penalty

Akaike Information Criterion (AIC)

AIC(S) = ℓn(�βS, �σ2
S) -k

Bayesian Information Criterion (BIC)

BIC(S) = ℓn(�βS, �σ2
S) -k

2 log n

Validation and training

�RV (S) =

m
�

i=1
(�Y ∗
i (S) -Y ∗
i )2
m = |{validation data}|, often n

4 or n

2

Leave-one-out cross-validation

�RCV (S) =

n
�

i=1
(Yi -�Y(i))2 =

n
�

i=1

�
Yi -�Yi(S)
1 -Uii(S)

�2

U(S) = XS(XT
S XS)-1XS ("hat matrix")

## 19 Non-parametric Function Estimation

### 19.1 Density Estimation

Estimate f(x), where f(x) = P [X ∈A] =
�

A f(x) dx.
Integrated square error (ise)

L(f, �fn) =
��
f(x) -�fn(x)
�2
dx = J(h) +
�
f 2(x) dx

Frequentist risk

R(f, �fn) = E
�
L(f, �fn)
�
=
�
b2(x) dx +
�
v(x) dx

b(x) = E
�
�fn(x)
�
-f(x)

v(x) = V
�
�fn(x)
�

#### 19.1.1 Histograms

Definitions

• Number of bins m
• Binwidth h = 1
m
• Bin Bj has νj observations
• Define �pj = νj/n and pj =
�

Bj f(u) du

Histogram estimator

�fn(x) =

m
�

j=1

�pj
h I(x ∈Bj)

E
�
�fn(x)
�
= pj

h

V
�
�fn(x)
�
= pj(1 -pj)

nh2

R( �fn, f) ≈h2

12

�
(f ′(u))2 du + 1

nh

h∗=
1
n1/3

�
6
�
(f ′(u))2 du

�1/3

R∗( �fn, f) ≈
C
n2/3
C =
�3
4

�2/3 ��
(f ′(u))2 du
�1/3

Cross-validation estimate of E [J(h)]

�JCV (h) =
�
�f 2
n(x) dx -2

n

n
�

i=1
�f(-i)(Xi) =
2
(n -1)h n + 1
(n -1)h

m
�

j=1
�p2
j

#### 19.1.2 Kernel Density Estimator (KDE)

Kernel K

• K(x) ≥0
•
�
K(x) dx = 1
•
�
xK(x) dx = 0
•
�
x2K(x) dx ≡σ2
K > 0

KDE

�fn(x) = 1
n

n
�

i=1

1
hK
�x -Xi
h

�

R(f, �fn) ≈1

4(hσK)4
�
(f ′′(x))2 dx + 1

nh

�
K2(x) dx

h∗= c-2/5
1
c-1/5
2
c-1/5
3
n1/5
c1 = σ2
K, c2 =
�
K2(x) dx, c3 =
�
(f ′′(x))2 dx

R∗(f, �fn) =
c4
n4/5
c4 = 5

4(σ2
K)2/5
��
K2(x) dx
�4/5

�
��
�
C(K)

��
(f ′′)2 dx
�1/5

Epanechnikov Kernel

K(x) =

�
3
4
√

5(1-x2/5)
|x| <
√

5

0
otherwise

Cross-validation estimate of E [J(h)]

�JCV (h) =
�
�f 2
n(x) dx -2

n

n
�

i=1
�f(-i)(Xi) ≈
1
hn2

n
�

i=1

n
�

j=1
K∗
�Xi -Xj
h

�
+ 2

nhK(0)

K∗(x) = K(2)(x) -2K(x)
K(2)(x) =
�
K(x -y)K(y) dy

### 19.2 Non-parametric Regression

Estimate f(x) where f(x) = E [Y | X = x]. Consider pairs of points
(x1, Y1), . . . , (xn, Yn) related by

Yi = r(xi) + ϵi
E [ϵi] = 0

V [ϵi] = σ2

k-nearest Neighbor Estimator

�r(x) = 1
k

�

i:xi∈Nk(x)
Yi
where Nk(x) = {k values of x1, . . . , xn closest to x}

Nadaraya-Watson Kernel Estimator

�r(x) =

n
�

i=1
wi(x)Yi

wi(x) =
K
�x-xi
h
�

�n
j=1 K
�
x-xj

h
�
∈[0, 1]

R(�rn, r) ≈h4

4

��
x2K2(x) dx
�4 ��
r′′(x) + 2r′(x)f ′(x)

f(x)

�2
dx

+
�σ2 �
K2(x) dx
nhf(x)
dx

h∗≈
c1
n1/5

R∗(�rn, r) ≈
c2
n4/5

Cross-validation estimate of E [J(h)]

�JCV (h) =

n
�

i=1
(Yi -�r(-i)(xi))2 =

n
�

i=1

(Yi -�r(xi))2
�

1 -
K(0)
�n
j=1 K
�x-xj
h
�

�2

### 19.3 Smoothing Using Orthogonal Functions

Approximation

r(x) =

∞
�

j=1
βjφj(x) ≈

J
�

i=1
βjφj(x)

Multivariate regression
Y = Φβ + η

where
ηi = ϵi
and
Φ =

φ0(x1)
· · ·
φJ(x1)
...
...
...
φ0(xn)
· · ·
φJ(xn)

Least squares estimator

�β = (ΦT Φ)-1ΦT Y

≈1

nΦT Y
(for equally spaced observations only)

Cross-validation estimate of E [J(h)]

�RCV (J) =

n
�

i=1

Yi -

J
�

j=1
φj(xi)�βj,(-i)

2

## 20 Stochastic Processes

Stochastic Process

{Xt : t ∈T}
T =

�
{0, ±1, . . . } = Z
discrete
[0, ∞)
continuous

• Notations Xt, X(t)
• State space X
• Index set T

### 20.1 Markov Chains

Markov chain

P [Xn = x | X0, . . . , Xn-1] = P [Xn = x | Xn-1]
∀n ∈T, x ∈X

Transition probabilities

pij ≡P [Xn+1 = j | Xn = i]

pij(n) ≡P [Xm+n = j | Xm = i]
n-step

Transition matrix P (n-step: Pn)

• (i, j) element is pij
• pij > 0
• �

i pij = 1

Chapman-Kolmogorov

pij(m + n) =
�

k
pij(m)pkj(n)

Pm+n = PmPn

Pn = P × · · · × P = Pn

Marginal probability

µn = (µn(1), . . . , µn(N))
where
µi(i) = P [Xn = i]

µ0 ≜initial distribution

µn = µ0Pn

### 20.2 Poisson Processes

Poisson process

• {Xt : t ∈[0, ∞)} = number of events up to and including time t
• X0 = 0
• Independent increments:

∀t0 < · · · < tn : Xt1 -Xt0 ⊥⊥· · · ⊥⊥Xtn -Xtn-1

• Intensity function λ(t)

– P [Xt+h -Xt = 1] = λ(t)h + o(h)
– P [Xt+h -Xt = 2] = o(h)

• Xs+t -Xs ∼Po (m(s + t) -m(s)) where m(t) =
�t
0 λ(s) ds

Homogeneous Poisson process

λ(t) ≡λ =⇒Xt ∼Po (λt)
λ > 0

Waiting times

Wt := time at which Xt occurs

Wt ∼Gamma
�
t, 1

λ

�

Interarrival times

St = Wt+1 -Wt

St ∼Exp
�1
λ

�

t
Wt-1
Wt

St

## 21 Time Series

Mean function

µxt = E [xt] =
�∞

-∞
xft(x) dx

Autocovariance function

γx(s, t) = E [(xs -µs)(xt -µt)] = E [xsxt] -µsµt

γx(t, t) = E
�
(xt -µt)2�
= V [xt]

Autocorrelation function (ACF)

ρ(s, t) =
Cov [xs, xt]
�
V [xs] V [xt]
=
γ(s, t)
�
γ(s, s)γ(t, t)

Cross-covariance function (CCV)

γxy(s, t) = E [(xs -µxs)(yt -µyt)]

Cross-correlation function (CCF)

ρxy(s, t) =
γxy(s, t)
�
γx(s, s)γy(t, t)

Backshift operator
Bk(xt) = xt-k

Difference operator
∇d = (1 -B)d

White noise

• wt ∼wn(0, σ2
w)

• Gaussian: wt
iid
∼N
�
0, σ2
w
�

• E [wt] = 0
t ∈T
• V [wt] = σ2
t ∈T
• γw(s, t) = 0
s ̸= t ∧s, t ∈T

Random walk

• Drift δ
• xt = δt + �t
j=1 wj
• E [xt] = δt

Symmetric moving average

mt =

k
�

j=-k
ajxt-j
where aj = a-j ≥0 and

k
�

j=-k
aj = 1

### 21.1 Stationary Time Series

Strictly stationary

P [xt1 ≤c1, . . . , xtk ≤ck] = P [xt1+h ≤c1, . . . , xtk+h ≤ck]

∀k ∈N, tk, ck, h ∈Z

Weakly stationary

• E
�
x2
t
�
< ∞
∀t ∈Z
• E
�
x2
t
�
= m
∀t ∈Z
• γx(s, t) = γx(s + r, t + r)
∀r, s, t ∈Z

Autocovariance function

• γ(h) = E [(xt+h -µ)(xt -µ)]
∀h ∈Z
• γ(0) = E
�
(xt -µ)2�

• γ(0) ≥0
• γ(0) ≥|γ(h)|
• γ(h) = γ(-h)

Autocorrelation function (ACF)

ρx(h) =
Cov [xt+h, xt]
�
V [xt+h] V [xt]
=
γ(t + h, t)
�
γ(t + h, t + h)γ(t, t)
= γ(h)

γ(0)

Jointly stationary time series

γxy(h) = E [(xt+h -µx)(yt -µy)]

ρxy(h) =
γxy(h)
�
γx(0)γy(h)

Linear process

xt = µ +

∞
�

j=-∞
ψjwt-j
where

∞
�

j=-∞
|ψj| < ∞

γ(h) = σ2
w

∞
�

j=-∞
ψj+hψj

### 21.2 Estimation of Correlation

Sample mean

¯x = 1

n

n
�

t=1
xt

Sample variance

V [¯x] = 1

n

n
�

h=-n

�
1 -|h|

n

�
γx(h)

Sample autocovariance function

�γ(h) = 1
n

n-h
�

t=1
(xt+h -¯x)(xt -¯x)

Sample autocorrelation function

�ρ(h) = �γ(h)
�γ(0)

Sample cross-variance function

�γxy(h) = 1
n

n-h
�

t=1
(xt+h -¯x)(yt -y)

Sample cross-correlation function

�ρxy(h) =
�γxy(h)
�
�γx(0)�γy(0)

Properties

• σ�ρx(h) =
1
√n if xt is white noise

• σ�ρxy(h) =
1
√n if xt or yt is white noise

### 21.3 Non-Stationary Time Series

Classical decomposition model

xt = µt + st + wt

• µt = trend
• st = seasonal component
• wt = random noise term

#### 21.3.1 Detrending

Least squares

1. Choose trend model, e.g., µt = β0 + β1t + β2t2

2. Minimize rss to obtain trend estimate �µt = �β0 + �β1t + �β2t2

3. Residuals ≜noise wt

Moving average

• The low-pass filter vt is a symmetric moving average mt with aj =
1
2k+1:

vt =
1
2k + 1

k
�

i=-k
xt-1

• If
1
2k+1
�k
i=-k wt-j ≈0, a linear trend function µt = β0 + β1t passes
without distortion

Differencing

• µt = β0 + β1t =⇒∇xt = β1

### 21.4 ARIMA models

Autoregressive polynomial

φ(z) = 1 -φ1z -· · · -φpzp
z ∈C ∧φp ̸= 0

Autoregressive operator

φ(B) = 1 -φ1B -· · · -φpBp

Autoregressive model order p, AR (p)

xt = φ1xt-1 + · · · + φpxt-p + wt ⇐⇒φ(B)xt = wt

AR (1)

• xt = φk(xt-k) +

k-1
�

j=0
φj(wt-j)
k→∞,|φ|<1
=

∞
�

j=0
φj(wt-j)

�
��
�
linear process
• E [xt] = �∞
j=0 φj(E [wt-j]) = 0

• γ(h) = Cov [xt+h, xt] = σ2
wφh

1-φ2

• ρ(h) = γ(h)
γ(0) = φh

• ρ(h) = φρ(h -1)
h = 1, 2, . . .

Moving average polynomial

θ(z) = 1 + θ1z + · · · + θqzq
z ∈C ∧θq ̸= 0

Moving average operator

θ(B) = 1 + θ1B + · · · + θpBp

MA (q) (moving average model order q)

xt = wt + θ1wt-1 + · · · + θqwt-q ⇐⇒xt = θ(B)wt

E [xt] =

q
�

j=0
θjE [wt-j] = 0

γ(h) = Cov [xt+h, xt] =

�
σ2
w
�q-h
j=0 θjθj+h
0 ≤h ≤q
0
h > q

MA (1)
xt = wt + θwt-1

γ(h) =

(1 + θ2)σ2
w
h = 0
θσ2
w
h = 1
0
h > 1

ρ(h) =

�
θ
(1+θ2)
h = 1

0
h > 1

ARMA (p, q)

xt = φ1xt-1 + · · · + φpxt-p + wt + θ1wt-1 + · · · + θqwt-q

φ(B)xt = θ(B)wt
Partial autocorrelation function (PACF)

• xh-1
i
≜regression of xi on {xh-1, xh-2, . . . , x1}
• φhh = corr(xh -xh-1
h
, x0 -xh-1
0
)
h ≥2
• E.g., φ11 = corr(x1, x0) = ρ(1)

ARIMA (p, d, q)
∇dxt = (1 -B)dxt is ARMA (p, q)

φ(B)(1 -B)dxt = θ(B)wt
Exponentially Weighted Moving Average (EWMA)

xt = xt-1 + wt -λwt-1

xt =

∞
�

j=1
(1 -λ)λj-1xt-j + wt
when |λ| < 1

˜xn+1 = (1 -λ)xn + λ˜xn

Seasonal ARIMA

• Denoted by ARIMA (p, d, q) × (P, D, Q)s
• ΦP (Bs)φ(B)∇D
s ∇dxt = δ + ΘQ(Bs)θ(B)wt

#### 21.4.1 Causality and Invertibility

ARMA (p, q) is causal (future-independent) ⇐⇒∃{ψj} : �∞
j=0 ψj < ∞such that

xt =

∞
�

j=0
wt-j = ψ(B)wt

ARMA (p, q) is invertible ⇐⇒∃{πj} : �∞
j=0 πj < ∞such that

π(B)xt =

∞
�

j=0
Xt-j = wt

Properties

• ARMA (p, q) causal ⇐⇒
roots of φ(z) lie outside the unit circle

ψ(z) =

∞
�

j=0
ψjzj = θ(z)

φ(z)
|z| ≤1

• ARMA (p, q) invertible ⇐⇒roots of θ(z) lie outside the unit circle

π(z) =

∞
�

j=0
πjzj = φ(z)

θ(z)
|z| ≤1

Behavior of the ACF and PACF for causal and invertible ARMA models

AR (p)
MA (q)
ARMA (p, q)
ACF
tails off
cuts off after lag q
tails off
PACF
cuts off after lag p
tails off q
tails off

### 21.5 Spectral Analysis

Periodic process

xt = A cos(2πωt + φ)

= U1 cos(2πωt) + U2 sin(2πωt)

• Frequency index ω (cycles per unit time), period 1/ω

• Amplitude A
• Phase φ
• U1 = A cos φ and U2 = A sin φ often normally distributed rv's

Periodic mixture

xt =

q
�

k=1
(Uk1 cos(2πωkt) + Uk2 sin(2πωkt))

• Uk1, Uk2, for k = 1, . . . , q, are independent zero-mean rv's with variances σ2
k
• γ(h) = �q
k=1 σ2
k cos(2πωkh)
• γ(0) = E
�
x2
t
�
= �q
k=1 σ2
k
Spectral representation of a periodic process

γ(h) = σ2 cos(2πω0h)

= σ2

2 e-2πiω0h + σ2

2 e2πiω0h

=
�1/2

-1/2
e2πiωh dF(ω)

Spectral distribution function

F(ω) =

0
ω < -ω0
σ2/2
-ω ≤ω < ω0
σ2
ω ≥ω0

• F(-∞) = F(-1/2) = 0
• F(∞) = F(1/2) = γ(0)

Spectral density

f(ω) =

∞
�

h=-∞
γ(h)e-2πiωh
-1

2 ≤ω ≤1

2

• Needs �∞
h=-∞|γ(h)| < ∞=⇒γ(h) =
�1/2
-1/2 e2πiωhf(ω) dω
h = 0, ±1, . . .

• f(ω) ≥0
• f(ω) = f(-ω)
• f(ω) = f(1 -ω)

• γ(0) = V [xt] =
�1/2
-1/2 f(ω) dω

• White noise: fw(ω) = σ2
w
• ARMA (p, q) , φ(B)xt = θ(B)wt:

fx(ω) = σ2
w
|θ(e-2πiω)|2

|φ(e-2πiω)|2

where φ(z) = 1 -�p
k=1 φkzk and θ(z) = 1 + �q
k=1 θkzk

Discrete Fourier Transform (DFT)

d(ωj) = n-1/2
n
�

i=1
xte-2πiωjt

Fourier/Fundamental frequencies

ωj = j/n

Inverse DFT

xt = n-1/2
n-1
�

j=0
d(ωj)e2πiωjt

Periodogram
I(j/n) = |d(j/n)|2

Scaled Periodogram

P(j/n) = 4

nI(j/n)

=

�
2
n

n
�

t=1
xt cos(2πtj/n

�2

+

�
2
n

n
�

t=1
xt sin(2πtj/n

�2

## 22 Math

### 22.1 Gamma Function

• Ordinary: Γ(s) =
�∞

0
ts-1e-tdt

• Upper incomplete: Γ(s, x) =
�∞

x
ts-1e-tdt

• Lower incomplete: γ(s, x) =
�x

0
ts-1e-tdt

• Γ(α + 1) = αΓ(α)
α > 1
• Γ(n) = (n -1)!
n ∈N
• Γ(1/2) = √π

### 22.2 Beta Function

• Ordinary: B(x, y) = B(y, x) =
�1

0
tx-1(1 -t)y-1 dt = Γ(x)Γ(y)

Γ(x + y)

• Incomplete: B(x; a, b) =
�x

0
ta-1(1 -t)b-1 dt

• Regularized incomplete:

Ix(a, b) = B(x; a, b)

B(a, b)

a,b∈N
=

a+b-1
�

j=a

(a + b -1)!
j!(a + b -1 -j)!xj(1 -x)a+b-1-j

• I0(a, b) = 0
I1(a, b) = 1
• Ix(a, b) = 1 -I1-x(b, a)

### 22.3 Series

Finite

•

n
�

k=1
k = n(n + 1)

2

•

n
�

k=1
(2k -1) = n2

•

n
�

k=1
k2 = n(n + 1)(2n + 1)

6

•

n
�

k=1
k3 =
�n(n + 1)
2

�2

•

n
�

k=0
ck = cn+1 -1

c -1
c ̸= 1

Binomial

•

n
�

k=0

�n
k

�
= 2n

•

n
�

k=0

�r + k
k

�
=
�r + n + 1
n

�

•

n
�

k=0

�k
m

�
=
�n + 1
m + 1

�

• Vandermonde's Identity:
r
�

k=0

�m
k

��
n
r -k

�
=
�m + n
r

�

• Binomial Theorem:
n
�

k=0

�n
k

�
an-kbk = (a + b)n

Infinite

•

∞
�

k=0
pk =
1
1 -p,

∞
�

k=1
pk =
p
1 -p
|p| < 1

•

∞
�

k=0
kpk-1 = d

dp

�∞
�

k=0
pk
�

= d

dp

�
1
1 -p

�
=
1
(1 -p)2
|p| < 1

•

∞
�

k=0

�r + k -1
k

�
xk = (1 -x)-r
r ∈N+

•

∞
�

k=0

�α
k

�
pk = (1 + p)α
|p| < 1 , α ∈C

### 22.4 Combinatorics

Sampling

k out of n
w/o replacement
w/ replacement

ordered
nk =

k-1
�

i=0
(n -i) =
n!
(n -k)!
nk

unordered

�n
k

�
= nk

k! =
n!
k!(n -k)!

�n -1 + r
r

�
=
�n -1 + r
n -1

�

Stirling numbers, 2nd kind
�n
k

�
= k
�n -1
k

�
+
�n -1
k -1

�
1 ≤k ≤n
�n
0

�
=

�
1
n = 0
0
else

Partitions

Pn+k,k =

n
�

i=1
Pn,i
k > n : Pn,k = 0
n ≥1 : Pn,0 = 0, P0,0 = 1

Balls and Urns
f : B →U
D = distinguishable, ¬D = indistinguishable.

|B| = n, |U| = m
f arbitrary
f injective
f surjective
f bijective

B : D, U : ¬D
mn

�
mn
m ≥n
0
else
m!
�n
m

�
�
n!
m = n
0
else

B : ¬D, U : D

�n + n -1
n

�
�m
n

�
�n -1
m -1

�
�
1
m = n
0
else

B : D, U : ¬D

m
�

k=1

�n
k

�
�
1
m ≥n
0
else

�n
m

�
�
1
m = n
0
else

B : ¬D, U : ¬D

m
�

k=1
Pn,k

�
1
m ≥n
0
else
Pn,m

�
1
m = n
0
else

References

[1] P. G. Hoel, S. C. Port, and C. J. Stone. Introduction to Probability Theory.
Brooks Cole, 1972.

[2] L. M. Leemis and J. T. McQueston. Univariate Distribution Relationships.
The American Statistician, 62(1):45–53, 2008.

[3] R. H. Shumway and D. S. Stoffer. Time Series Analysis and Its Applications
With R Examples. Springer, 2006.

[4] A. Steger. Diskrete Strukturen – Band 1: Kombinatorik, Graphentheorie,
Algebra. Springer, 2001.

[5] A. Steger. Diskrete Strukturen – Band 2: Wahrscheinlichkeitstheorie und
Statistik. Springer, 2002.

[6] L. Wasserman. All of Statistics: A Concise Course in Statistical Inference.
Springer, 2003.