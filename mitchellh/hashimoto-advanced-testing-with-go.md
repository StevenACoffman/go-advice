
hello and uh thanks for being here hello and uh thanks for being here
  hello and uh thanks for being here
there's a a lot more people than there's a a lot more people than
  there's a a lot more people than
gophercon in 2014 uh like an order gophercon in 2014 uh like an order
  gophercon in 2014 uh like an order
magnitude I feel like in this magnitude I feel like in this
  magnitude I feel like in this
room um today I want to talk about room um today I want to talk about
  room um today I want to talk about
Advanced testing with go um I got a good Advanced testing with go um I got a good
  Advanced testing with go um I got a good
introduction so I could skip this slide introduction so I could skip this slide
  introduction so I could skip this slide
with my face on it um I founded a with my face on it um I founded a
  with my face on it um I founded a
company called Hashi Corp uh thank you company called Hashi Corp uh thank you
  company called Hashi Corp uh thank you
Ashley for this really adorable uh Ashley for this really adorable uh
  Ashley for this really adorable uh
graphic um my fiance just is in love graphic um my fiance just is in love
  graphic um my fiance just is in love
with this thing um but we love go at with this thing um but we love go at
  with this thing um but we love go at
Hashi Corp a majority of our lines of Hashi Corp a majority of our lines of
  Hashi Corp a majority of our lines of
code and projects kind of like by any code and projects kind of like by any
  code and projects kind of like by any
metric is written uh in go and uh it was metric is written uh in go and uh it was
  metric is written uh in go and uh it was
mentioned that we have these are the mentioned that we have these are the
  mentioned that we have these are the
open source projects we have that are open source projects we have that are
  open source projects we have that are
written in go or no these are the open written in go or no these are the open
  written in go or no these are the open
source projects we have that are written source projects we have that are written
  source projects we have that are written
in go um but this doesn't cover all the in go um but this doesn't cover all the
  in go um but this doesn't cover all the
Enterprise work we've done uh Clos Enterprise work we've done uh Clos
  Enterprise work we've done uh Clos
Source sort of libraries we use for Source sort of libraries we use for
  Source sort of libraries we use for
those Enterprise products on a lot more those Enterprise products on a lot more
  those Enterprise products on a lot more
written and go we we write go all the written and go we we write go all the
  written and go we we write go all the
time um and something that makes this a time um and something that makes this a
  time um and something that makes this a
little bit unique um so Go's been our little bit unique um so Go's been our
  little bit unique um so Go's been our
primary language for 5 years uh it was primary language for 5 years uh it was
  primary language for 5 years uh it was
definitely a bet five years ago uh and definitely a bet five years ago uh and
  definitely a bet five years ago uh and
it's paid off it's been great uh our it's paid off it's been great uh our
  it's paid off it's been great uh our
projects have a bunch of properties that projects have a bunch of properties that
  projects have a bunch of properties that
make testing rather interesting so make testing rather interesting so
  make testing rather interesting so
there's the scale aspect which a lot of there's the scale aspect which a lot of
  there's the scale aspect which a lot of
people have um our projects are deployed people have um our projects are deployed
  people have um our projects are deployed
by millions of units whatever that unit by millions of units whatever that unit
  by millions of units whatever that unit
ends up being there's millions of users ends up being there's millions of users
  ends up being there's millions of users
for sure um at this point we're sort of for sure um at this point we're sort of
  for sure um at this point we're sort of
hitting the millions of servers as well hitting the millions of servers as well
  hitting the millions of servers as well
uh and then there's it's deployed uh and then there's it's deployed
  uh and then there's it's deployed
significantly uh in Enterprises which significantly uh in Enterprises which
  significantly uh in Enterprises which
has a different expectation of software has a different expectation of software
  has a different expectation of software
uh in a variety of ways so there's that uh in a variety of ways so there's that
  uh in a variety of ways so there's that
aspect uh we build a a bunch of aspect uh we build a a bunch of
  aspect uh we build a a bunch of
different categories of software that different categories of software that
  different categories of software that
differ product by product which affects differ product by product which affects
  differ product by product which affects
our approach to testing and our our approach to testing and our
  our approach to testing and our
viewpoint on testing so we have a number viewpoint on testing so we have a number
  viewpoint on testing so we have a number
of distributed systems uh console surf of distributed systems uh console surf
  of distributed systems uh console surf
nomad these could be viewed as you know nomad these could be viewed as you know
  nomad these could be viewed as you know
sort of at their core distributed sort of at their core distributed
  sort of at their core distributed
systems we have certain software that systems we have certain software that
  systems we have certain software that
has performance that matters a lot um in has performance that matters a lot um in
  has performance that matters a lot um in
C by certain metrics and so console uh C by certain metrics and so console uh
  C by certain metrics and so console uh
read performance matters quite a bit uh read performance matters quite a bit uh
  read performance matters quite a bit uh
Nomad scheduling performance matters a Nomad scheduling performance matters a
  Nomad scheduling performance matters a
lot um we have tools like Vault uh vault lot um we have tools like Vault uh vault
  lot um we have tools like Vault uh vault
is a security tool and so it has a high is a security tool and so it has a high
  is a security tool and so it has a high
you know it almost has to be sort of you know it almost has to be sort of
  you know it almost has to be sort of
perfect degree of security we need to perfect degree of security we need to
  perfect degree of security we need to
maintain maintain
  maintain
um and testing has to come into that in um and testing has to come into that in
  um and testing has to come into that in
addition to other things uh and then at addition to other things uh and then at
  addition to other things uh and then at
the end there's there's correctness of the end there's there's correctness of
  the end there's there's correctness of
when things don't work correctly it when things don't work correctly it
  when things don't work correctly it
could be very detrimental so all could be very detrimental so all
  could be very detrimental so all
software has bugs we ship bugs everyone software has bugs we ship bugs everyone
  software has bugs we ship bugs everyone
ships bugs I'm not saying we ship ships bugs I'm not saying we ship
  ships bugs I'm not saying we ship
perfect code we don't um but things like perfect code we don't um but things like
  perfect code we don't um but things like
terraform we need to make sure that when terraform we need to make sure that when
  terraform we need to make sure that when
we ship a terraform update someone we ship a terraform update someone
  we ship a terraform update someone
doesn't run a change and their whole doesn't run a change and their whole
  doesn't run a change and their whole
infrastructure disappears um or someone infrastructure disappears um or someone
  infrastructure disappears um or someone
updates console and they lose all their updates console and they lose all their
  updates console and they lose all their
data um or or someone updates Nomad and data um or or someone updates Nomad and
  data um or or someone updates Nomad and
we decide to deschedule all their we decide to deschedule all their
  we decide to deschedule all their
services there's catastrophic things uh services there's catastrophic things uh
  services there's catastrophic things uh
on the correctness scale that we need to on the correctness scale that we need to
  on the correctness scale that we need to
definitely prevent uh and then coming definitely prevent uh and then coming
  definitely prevent uh and then coming
back from the catastrophic things back from the catastrophic things
  back from the catastrophic things
there's just uh more details of there's just uh more details of
  there's just uh more details of
correctness we we want to correctness we we want to
  correctness we we want to
ensure so the ways Talk Works um is that ensure so the ways Talk Works um is that
  ensure so the ways Talk Works um is that
there's there's really two parts to there's there's really two parts to
  there's there's really two parts to
testing um and you need both of them to testing um and you need both of them to
  testing um and you need both of them to
to produce uh to produce good testing to produce uh to produce good testing
  to produce uh to produce good testing
and so there's test methodologies uh and so there's test methodologies uh
  and so there's test methodologies uh
which is writing the tests themselves which is writing the tests themselves
  which is writing the tests themselves
and how to write the tests uh and then and how to write the tests uh and then
  and how to write the tests uh and then
there's writing testable code which I there's writing testable code which I
  there's writing testable code which I
think is equally important you can't think is equally important you can't
  think is equally important you can't
just take any kind of code and write just take any kind of code and write
  just take any kind of code and write
great tests for it and so this is also great tests for it and so this is also
  great tests for it and so this is also
the slide style um to make it really the slide style um to make it really
  the slide style um to make it really
simple so the slides that are in black simple so the slides that are in black
  simple so the slides that are in black
background um are going to be about test background um are going to be about test
  background um are going to be about test
methodologies and the slides that are in methodologies and the slides that are in
  methodologies and the slides that are in
a white background are going to be about a white background are going to be about
  a white background are going to be about
writing testable code writing testable code
  writing testable code
um so to explain that in a little bit um so to explain that in a little bit
  um so to explain that in a little bit
more detail so test methodology starting more detail so test methodology starting
  more detail so test methodology starting
with the slide styling right away um are with the slide styling right away um are
  with the slide styling right away um are
methods to spe test specific types of methods to spe test specific types of
  methods to spe test specific types of
cases that you see uh when you're when cases that you see uh when you're when
  cases that you see uh when you're when
you're R tests um their techniques to you're R tests um their techniques to
  you're R tests um their techniques to
write better tests um better is defined write better tests um better is defined
  write better tests um better is defined
in a bunch of different ways throughout in a bunch of different ways throughout
  in a bunch of different ways throughout
the talk um and it's it's trying to the talk um and it's it's trying to
  the talk um and it's it's trying to
explain that there's a lot more to explain that there's a lot more to
  explain that there's a lot more to
testing as we all probably know than testing as we all probably know than
  testing as we all probably know than
assert something assert something
  assert something
and then there's writing testable code and then there's writing testable code
  and then there's writing testable code
um which like I said is how to write um which like I said is how to write
  um which like I said is how to write
test how to write code that could be test how to write code that could be
  test how to write code that could be
tested well and easily um and this just tested well and easily um and this just
  tested well and easily um and this just
as important um it's very common uh both as important um it's very common uh both
  as important um it's very common uh both
from Junior to senior Engineers I see it from Junior to senior Engineers I see it
  from Junior to senior Engineers I see it
all the time to hear things hear hear all the time to hear things hear hear
  all the time to hear things hear hear
people say this just can't be tested people say this just can't be tested
  people say this just can't be tested
well so I didn't write a test but I ran well so I didn't write a test but I ran
  well so I didn't write a test but I ran
through it or something um and they through it or something um and they
  through it or something um and they
might not be wrong like they they I might not be wrong like they they I
  might not be wrong like they they I
might look at the code and say like might look at the code and say like
  might look at the code and say like
you're not wrong this can't be tested in you're not wrong this can't be tested in
  you're not wrong this can't be tested in
any reasonable way um but by refactoring any reasonable way um but by refactoring
  any reasonable way um but by refactoring
in a certain way by rearching it in a in a certain way by rearching it in a
  in a certain way by rearching it in a
certain way you could usually get to a certain way you could usually get to a
  certain way you could usually get to a
point where at least 90% of that point where at least 90% of that
  point where at least 90% of that
functionality is tested and maybe the functionality is tested and maybe the
  functionality is tested and maybe the
10% that's very hard to test is remains 10% that's very hard to test is remains
  10% that's very hard to test is remains
out there um and so at Hashi Corp we uh out there um and so at Hashi Corp we uh
  out there um and so at Hashi Corp we uh
I don't think we've ever seen anything I don't think we've ever seen anything
  I don't think we've ever seen anything
that can't be tested very well and we that can't be tested very well and we
  that can't be tested very well and we
hit a pretty broad spectrum um and hit a pretty broad spectrum um and
  hit a pretty broad spectrum um and
rewriting existing code could be a pain rewriting existing code could be a pain
  rewriting existing code could be a pain
but it usually for testability it's but it usually for testability it's
  but it usually for testability it's
worth worth
  worth
it okay so from here on out we're just it okay so from here on out we're just
  it okay so from here on out we're just
going to dive right into it and it's going to dive right into it and it's
  going to dive right into it and it's
just going to be um I mean it's just just going to be um I mean it's just
  just going to be um I mean it's just
going to be like a machine gun of about going to be like a machine gun of about
  going to be like a machine gun of about
like 30 I think 30 different methods and like 30 I think 30 different methods and
  like 30 I think 30 different methods and
uh how to write testable code in total uh how to write testable code in total
  uh how to write testable code in total
and so we're just we're just going to and so we're just we're just going to
  and so we're just we're just going to
get going oh the last thing in the slide get going oh the last thing in the slide
  get going oh the last thing in the slide
format is they are in order roughly from format is they are in order roughly from
  format is they are in order roughly from
things I expect everyone in this room to things I expect everyone in this room to
  things I expect everyone in this room to
know to getting more and more esoteric know to getting more and more esoteric
  know to getting more and more esoteric
and weird so at the be don't be and weird so at the be don't be
  and weird so at the be don't be
discouraged at the beginning if you're discouraged at the beginning if you're
  discouraged at the beginning if you're
if you think oh this is just a really if you think oh this is just a really
  if you think oh this is just a really
beginner talk um I'm going to try to beginner talk um I'm going to try to
  beginner talk um I'm going to try to
ramp it up for you so um I won't promise ramp it up for you so um I won't promise
  ramp it up for you so um I won't promise
that everyone will take at least one that everyone will take at least one
  that everyone will take at least one
thing away um but I hope that everyone thing away um but I hope that everyone
  thing away um but I hope that everyone
will take at least one thing will take at least one thing
  will take at least one thing
away um so starting right at the away um so starting right at the
  away um so starting right at the
beginning with with simple stuff I think beginning with with simple stuff I think
  beginning with with simple stuff I think
um and skipping how to write a a single um and skipping how to write a a single
  um and skipping how to write a a single
go test I assume everyone knows how to go test I assume everyone knows how to
  go test I assume everyone knows how to
write one go test so skipping that um is write one go test so skipping that um is
  write one go test so skipping that um is
subtests um subtests are new in go 1.8 subtests um subtests are new in go 1.8
  subtests um subtests are new in go 1.8
um officially um and they look like this um officially um and they look like this
  um officially um and they look like this
so they let you write a test and they so they let you write a test and they
  so they let you write a test and they
let you um Nest sort of by calling run let you um Nest sort of by calling run
  let you um Nest sort of by calling run
they let you Nest subtest within a test they let you Nest subtest within a test
  they let you Nest subtest within a test
and oh that's good I didn't do the and oh that's good I didn't do the
  and oh that's good I didn't do the
output but um if you run the test you output but um if you run the test you
  output but um if you run the test you
can Target those subtests um when you can Target those subtests um when you
  can Target those subtests um when you
look at the output it lists all the look at the output it lists all the
  look at the output it lists all the
subtests uh and there's a bunch of subtests uh and there's a bunch of
  subtests uh and there's a bunch of
benefits to these things so for example benefits to these things so for example
  benefits to these things so for example
your subtests are a closure so now your subtests are a closure so now
  your subtests are a closure so now
defers work within those you could use defers work within those you could use
  defers work within those you could use
if you have a huge set of test cases and if you have a huge set of test cases and
  if you have a huge set of test cases and
you're opening files or making network you're opening files or making network
  you're opening files or making network
connections or something you could connections or something you could
  connections or something you could
actually run defers in these subtests actually run defers in these subtests
  actually run defers in these subtests
rather than uh accumulating them or like rather than uh accumulating them or like
  rather than uh accumulating them or like
in the past before this was an official in the past before this was an official
  in the past before this was an official
thing I would actually just make an thing I would actually just make an
  thing I would actually just make an
anonymous function that I called right anonymous function that I called right
  anonymous function that I called right
away just to get the defer functionality away just to get the defer functionality
  away just to get the defer functionality
um as part of it um and you can Nest um as part of it um and you can Nest
  um as part of it um and you can Nest
them them
  them
infinitely so it's built in to go um infinitely so it's built in to go um
  infinitely so it's built in to go um
you're allowed to Target subtests and you're allowed to Target subtests and
  you're allowed to Target subtests and
you could just continue to Nest them you could just continue to Nest them
  you could just continue to Nest them
further if necessary further if necessary
  further if necessary
um and it's hard to explain the value of um and it's hard to explain the value of
  um and it's hard to explain the value of
subtests without talking about table subtests without talking about table
  subtests without talking about table
driven tests since I think this is the driven tests since I think this is the
  driven tests since I think this is the
nine out of 10 use case for subtests uh nine out of 10 use case for subtests uh
  nine out of 10 use case for subtests uh
that I've that I've
  that I've
found so table driven tests uh look a found so table driven tests uh look a
  found so table driven tests uh look a
little bit like this and I've used the little bit like this and I've used the
  little bit like this and I've used the
subtest syntax here too they're a way to subtest syntax here too they're a way to
  subtest syntax here too they're a way to
build a a table of data and a table of build a a table of data and a table of
  build a a table of data and a table of
test cases within a single test um and test cases within a single test um and
  test cases within a single test um and
run through them so in this case we're run through them so in this case we're
  run through them so in this case we're
testing addition just as an example and testing addition just as an example and
  testing addition just as an example and
so there's a bunch of cases up here so there's a bunch of cases up here
  so there's a bunch of cases up here
where we specify the operand for A and B where we specify the operand for A and B
  where we specify the operand for A and B
and then the expected value uh by adding and then the expected value uh by adding
  and then the expected value uh by adding
them together and then you could Loop them together and then you could Loop
  them together and then you could Loop
over them run the subtests uh and make over them run the subtests uh and make
  over them run the subtests uh and make
it work and there's a few things I'm it work and there's a few things I'm
  it work and there's a few things I'm
showing here um so one naming the showing here um so one naming the
  showing here um so one naming the
subtest so in this case I name it by subtest so in this case I name it by
  subtest so in this case I name it by
just with the the value like what what's just with the the value like what what's
  just with the the value like what what's
actually being tested um that's actually being tested um that's
  actually being tested um that's
sometimes useful and then I'll show you sometimes useful and then I'll show you
  sometimes useful and then I'll show you
in a little bit another option uh and in a little bit another option uh and
  in a little bit another option uh and
then using the actual subtest itself and then using the actual subtest itself and
  then using the actual subtest itself and
so we use TBL driven tests a lot um we so we use TBL driven tests a lot um we
  so we use TBL driven tests a lot um we
use it a lot at Hashi Corp I like to use it a lot at Hashi Corp I like to
  use it a lot at Hashi Corp I like to
almost default in a lot of cases the TBL almost default in a lot of cases the TBL
  almost default in a lot of cases the TBL
driven tests even if I have a single driven tests even if I have a single
  driven tests even if I have a single
case um I like to just set it up the case um I like to just set it up the
  case um I like to just set it up the
structure because if I look at something structure because if I look at something
  structure because if I look at something
and I think yeah there's probably a and I think yeah there's probably a
  and I think yeah there's probably a
scenario where we want to test other scenario where we want to test other
  scenario where we want to test other
parameters here one day in the future parameters here one day in the future
  parameters here one day in the future
I'll just set it up from the beginning I'll just set it up from the beginning
  I'll just set it up from the beginning
um with a table um so I like to do that um with a table um so I like to do that
  um with a table um so I like to do that
it's low overhead to add a new test um it's low overhead to add a new test um
  it's low overhead to add a new test um
which is the best case when you're which is the best case when you're
  which is the best case when you're
fixing a bug or you find a bug um fixing a bug or you find a bug um
  fixing a bug or you find a bug um
creating a regression test uh and adding creating a regression test uh and adding
  creating a regression test uh and adding
the case is super super easy for the the case is super super easy for the
  the case is super super easy for the
developer um it makes testing exhaustive developer um it makes testing exhaustive
  developer um it makes testing exhaustive
scenarios both simple uh technically but scenarios both simple uh technically but
  scenarios both simple uh technically but
also visually simple like uh depending also visually simple like uh depending
  also visually simple like uh depending
on what you're testing with a table like on what you're testing with a table like
  on what you're testing with a table like
it's easy to see visually if you've ex it's easy to see visually if you've ex
  it's easy to see visually if you've ex
exhaustively tested sort of all the exhaustively tested sort of all the
  exhaustively tested sort of all the
edges uh of your function um let's see edges uh of your function um let's see
  edges uh of your function um let's see
uh yeah we do this pattern a lot so I uh yeah we do this pattern a lot so I
  uh yeah we do this pattern a lot so I
recommend just doing this wherever uh recommend just doing this wherever uh
  recommend just doing this wherever uh
it's it's a little bit of overhead a lot it's it's a little bit of overhead a lot
  it's it's a little bit of overhead a lot
of value I think in the long run the of value I think in the long run the
  of value I think in the long run the
other thing I would say is consider other thing I would say is consider
  other thing I would say is consider
naming the cases uh so here's another naming the cases uh so here's another
  naming the cases uh so here's another
example where we just put a name in the example where we just put a name in the
  example where we just put a name in the
test case in the table uh thing and then test case in the table uh thing and then
  test case in the table uh thing and then
when we run the subtest we actually use when we run the subtest we actually use
  when we run the subtest we actually use
the name there uh we used to this is the name there uh we used to this is
  the name there uh we used to this is
like along four years ago when we did like along four years ago when we did
  like along four years ago when we did
table tests we used to uh just generate table tests we used to uh just generate
  table tests we used to uh just generate
all the the subtest names at the time all the the subtest names at the time
  all the the subtest names at the time
they weren't real sub tests uh with just they weren't real sub tests uh with just
  they weren't real sub tests uh with just
the data um but as you get more complex the data um but as you get more complex
  the data um but as you get more complex
TBL driven tests that starts becoming TBL driven tests that starts becoming
  TBL driven tests that starts becoming
really un clear um and if you just do really un clear um and if you just do
  really un clear um and if you just do
like a loop with indexes the indexes are like a loop with indexes the indexes are
  like a loop with indexes the indexes are
unclear like some people will just do unclear like some people will just do
  unclear like some people will just do
the name as the index uh of of the the the name as the index uh of of the the
  the name as the index uh of of the the
slice and when it's like test failure slice and when it's like test failure
  slice and when it's like test failure
you know test index number 3014 failed you know test index number 3014 failed
  you know test index number 3014 failed
uh it's really hard to find 314 there's uh it's really hard to find 314 there's
  uh it's really hard to find 314 there's
definitely been times like in terraform definitely been times like in terraform
  definitely been times like in terraform
where we've just gone through a file in where we've just gone through a file in
  where we've just gone through a file in
like zero one two three four um and like zero one two three four um and
  like zero one two three four um and
that's usually when we start adding the that's usually when we start adding the
  that's usually when we start adding the
names to things so that was a mistake we names to things so that was a mistake we
  names to things so that was a mistake we
made early on made early on
  made early on
um and I want to add as a disclaimer um and I want to add as a disclaimer
  um and I want to add as a disclaimer
like I'm not trying to claim at any like I'm not trying to claim at any
  like I'm not trying to claim at any
point during this this talk that any of point during this this talk that any of
  point during this this talk that any of
these things I say are novel uh a lot these things I say are novel uh a lot
  these things I say are novel uh a lot
like table driven test the first place I like table driven test the first place I
  like table driven test the first place I
saw it was in the goh standard Library saw it was in the goh standard Library
  saw it was in the goh standard Library
um a lot of these things are from the um a lot of these things are from the
  um a lot of these things are from the
goh standard Library I've seen them goh standard Library I've seen them
  goh standard Library I've seen them
maybe in other projects but I just I maybe in other projects but I just I
  maybe in other projects but I just I
don't remember anymore so it's just the don't remember anymore so it's just the
  don't remember anymore so it's just the
a a
  a
list okay let's keep going um test list okay let's keep going um test
  list okay let's keep going um test
fixtures so uh test fixures so you could fixtures so uh test fixures so you could
  fixtures so uh test fixures so you could
do this you could ask access data using do this you could ask access data using
  do this you could ask access data using
test fixtures and you can notice here test fixtures and you can notice here
  test fixtures and you can notice here
that I'm just accessing data relative to that I'm just accessing data relative to
  that I'm just accessing data relative to
the current working directory in a test the current working directory in a test
  the current working directory in a test
fixures folder you can name it anything fixures folder you can name it anything
  fixures folder you can name it anything
you want um so little known thing for a you want um so little known thing for a
  you want um so little known thing for a
lot of beginners in go is that go test lot of beginners in go is that go test
  lot of beginners in go is that go test
will always set the current working will always set the current working
  will always set the current working
directory as the package being tested so directory as the package being tested so
  directory as the package being tested so
when you do go test period SL period when you do go test period SL period
  when you do go test period SL period
period period and you're testing all period period and you're testing all
  period period and you're testing all
your packages uh each time it goes into your packages uh each time it goes into
  your packages uh each time it goes into
a new package it sets the current a new package it sets the current
  a new package it sets the current
working directory to the package being working directory to the package being
  working directory to the package being
tested and this is really helpful uh tested and this is really helpful uh
  tested and this is really helpful uh
because it lets you use relative paths because it lets you use relative paths
  because it lets you use relative paths
to access data if you need to for your to access data if you need to for your
  to access data if you need to for your
tests um at Hashi Corp we use the name tests um at Hashi Corp we use the name
  tests um at Hashi Corp we use the name
test fixtures um for no specific reason test fixtures um for no specific reason
  test fixtures um for no specific reason
as the directory where we store our test as the directory where we store our test
  as the directory where we store our test
data um but it's very useful for things data um but it's very useful for things
  data um but it's very useful for things
like loading example configurations we like loading example configurations we
  like loading example configurations we
test most of our software against Real test most of our software against Real
  test most of our software against Real
example configurations so just writing example configurations so just writing
  example configurations so just writing
files as you would if you running it um files as you would if you running it um
  files as you would if you running it um
model data for web applications actual model data for web applications actual
  model data for web applications actual
fixtures there now binary data super fixtures there now binary data super
  fixtures there now binary data super
useful um we use it to test you know useful um we use it to test you know
  useful um we use it to test you know
certain uh like Nomad how it handles um certain uh like Nomad how it handles um
  certain uh like Nomad how it handles um
tar.gz or uh Docker images and things tar.gz or uh Docker images and things
  tar.gz or uh Docker images and things
like that um we we use test pictures for like that um we we use test pictures for
  like that um we we use test pictures for
this just keep going uh golden files so this just keep going uh golden files so
  this just keep going uh golden files so
this is something that's definitely I this is something that's definitely I
  this is something that's definitely I
saw in the in the standard Library I saw in the in the standard Library I
  saw in the in the standard Library I
remember that um it's and golden files remember that um it's and golden files
  remember that um it's and golden files
is what they called it I think in the in is what they called it I think in the in
  is what they called it I think in the in
the standard Library which is why I call the standard Library which is why I call
  the standard Library which is why I call
it this here um golden files are a way it this here um golden files are a way
  it this here um golden files are a way
to uh compare complex test output to a to uh compare complex test output to a
  to uh compare complex test output to a
file that has the expected result so um file that has the expected result so um
  file that has the expected result so um
the place in the standard lib where I the place in the standard lib where I
  the place in the standard lib where I
first saw this I'm not sure if it's the first saw this I'm not sure if it's the
  first saw this I'm not sure if it's the
only place that exists is to test Gump only place that exists is to test Gump
  only place that exists is to test Gump
um when when they test gof fump they run um when when they test gof fump they run
  um when when they test gof fump they run
gof fump and then they compare the gof fump and then they compare the
  gof fump and then they compare the
resulting bytes to a golden files resulting bytes to a golden files
  resulting bytes to a golden files
contents and they put this flag which is contents and they put this flag which is
  contents and they put this flag which is
really interesting if you put if you put really interesting if you put if you put
  really interesting if you put if you put
a global flag flag um in your test it a global flag flag um in your test it
  a global flag flag um in your test it
actually becomes available on go test actually becomes available on go test
  actually becomes available on go test
and so they put a flag update that when and so they put a flag update that when
  and so they put a flag update that when
you use the flag update uh it actually you use the flag update uh it actually
  you use the flag update uh it actually
updates all the golden files so your updates all the golden files so your
  updates all the golden files so your
test will always pass when you use the test will always pass when you use the
  test will always pass when you use the
update flag because all the files will update flag because all the files will
  update flag because all the files will
get uploaded uh updated but you could get uploaded uh updated but you could
  get uploaded uh updated but you could
use that to update the golden files and use that to update the golden files and
  use that to update the golden files and
it's really um a much better way to it's really um a much better way to
  it's really um a much better way to
compare lots of bytes than just putting compare lots of bytes than just putting
  compare lots of bytes than just putting
the btes in a constant or something in the btes in a constant or something in
  the btes in a constant or something in
the footer of the file or in the test the footer of the file or in the test
  the footer of the file or in the test
itself uh and so uh just run I'll just itself uh and so uh just run I'll just
  itself uh and so uh just run I'll just
run through this real quick since it is run through this real quick since it is
  run through this real quick since it is
a lot of code but it's actually usually a lot of code but it's actually usually
  a lot of code but it's actually usually
the pattern that golden files take more the pattern that golden files take more
  the pattern that golden files take more
or less um you usually have a table with or less um you usually have a table with
  or less um you usually have a table with
golden files that's usually part of the golden files that's usually part of the
  golden files that's usually part of the
thing it's either a table or a generated thing it's either a table or a generated
  thing it's either a table or a generated
table from a directory uh or something table from a directory uh or something
  table from a directory uh or something
you go through you do something to get you go through you do something to get
  you go through you do something to get
the actual bytes that's the actual the actual bytes that's the actual
  the actual bytes that's the actual
assignment there um you load the golden assignment there um you load the golden
  assignment there um you load the golden
file by usually the name you load the file by usually the name you load the
  file by usually the name you load the
golden file if the update flag was golden file if the update flag was
  golden file if the update flag was
specified you update the golden file specified you update the golden file
  specified you update the golden file
um and and don't judge me I'm ignoring um and and don't judge me I'm ignoring
  um and and don't judge me I'm ignoring
errors here because we have limited errors here because we have limited
  errors here because we have limited
space I would test the errors usually um space I would test the errors usually um
  space I would test the errors usually um
then you read the golden file and you then you read the golden file and you
  then you read the golden file and you
just do a bite SQL check and if it fails just do a bite SQL check and if it fails
  just do a bite SQL check and if it fails
uh you come up with some way to show a uh you come up with some way to show a
  uh you come up with some way to show a
diff of that so like the fump test for diff of that so like the fump test for
  diff of that so like the fump test for
example actually have a really nice diff example actually have a really nice diff
  example actually have a really nice diff
function to show the diff uh in a way function to show the diff uh in a way
  function to show the diff uh in a way
that isn't just here's bytes one and that isn't just here's bytes one and
  that isn't just here's bytes one and
here's bytes two and they just don't here's bytes two and they just don't
  here's bytes two and they just don't
match just find the difference um but match just find the difference um but
  match just find the difference um but
depending what I'm testing sometimes depending what I'm testing sometimes
  depending what I'm testing sometimes
that will be what I do um or I'll that will be what I do um or I'll
  that will be what I do um or I'll
actually do the diff 
  
and then when you run go test this is and then when you run go test this is
  and then when you run go test this is
what it looks like you could just what it looks like you could just
  what it looks like you could just
introduce Flags uh as you want um and I introduce Flags uh as you want um and I
  introduce Flags uh as you want um and I
guess I don't have another section that guess I don't have another section that
  guess I don't have another section that
explains that but you could introduce explains that but you could introduce
  explains that but you could introduce
Flags to the go test command and we use Flags to the go test command and we use
  Flags to the go test command and we use
that for things other than golden files that for things other than golden files
  that for things other than golden files
uh we use that in some tests I'm trying uh we use that in some tests I'm trying
  uh we use that in some tests I'm trying
to think to to disable certain types of to think to to disable certain types of
  to think to to disable certain types of
very side effective tests or very very side effective tests or very
  very side effective tests or very
expensive or very slow tests um you expensive or very slow tests um you
  expensive or very slow tests um you
could just introduce 
  
Flags uh and I think yeah I think I Flags uh and I think yeah I think I
  Flags uh and I think yeah I think I
covered all this I have a lot of these covered all this I have a lot of these
  covered all this I have a lot of these
bullet points to make it more friendly bullet points to make it more friendly
  bullet points to make it more friendly
for when I upload the for when I upload the
  for when I upload the
slides okay so we did a lot of test slides okay so we did a lot of test
  slides okay so we did a lot of test
methodology and now we're going to see methodology and now we're going to see
  methodology and now we're going to see
our first how to write testable code um our first how to write testable code um
  our first how to write testable code um
and there's going to be a lot more from and there's going to be a lot more from
  and there's going to be a lot more from
here on out so Global State I think this here on out so Global State I think this
  here on out so Global State I think this
is pretty obvious um but avoid it as is pretty obvious um but avoid it as
  is pretty obvious um but avoid it as
much as possible there's a lot of much as possible there's a lot of
  much as possible there's a lot of
reasons to do this of course um but in reasons to do this of course um but in
  reasons to do this of course um but in
the context of tests it's really the context of tests it's really
  the context of tests it's really
important to avoid it because it makes important to avoid it because it makes
  important to avoid it because it makes
your tests depend you know change your tests depend you know change
  your tests depend you know change
depending potentially Chang depending on depending potentially Chang depending on
  depending potentially Chang depending on
the order They're ran um it makes it the order They're ran um it makes it
  the order They're ran um it makes it
difficult to reason about you know the difficult to reason about you know the
  difficult to reason about you know the
full sort of inputs that are necessary full sort of inputs that are necessary
  full sort of inputs that are necessary
to affect the behavior of of some some to affect the behavior of of some some
  to affect the behavior of of some some
thing so instead of try instead of using thing so instead of try instead of using
  thing so instead of try instead of using
Global State um we try to make Global State um we try to make
  Global State um we try to make
whatever's Global a configuration option whatever's Global a configuration option
  whatever's Global a configuration option
U maybe we set up a constant that's the U maybe we set up a constant that's the
  U maybe we set up a constant that's the
default value in the global scope uh but default value in the global scope uh but
  default value in the global scope uh but
we still always try to make a we still always try to make a
  we still always try to make a
configurable thing um because you configurable thing um because you
  configurable thing um because you
usually want to modify it or twiddle it usually want to modify it or twiddle it
  usually want to modify it or twiddle it
uh for uh for
  uh for
tests so here's some examples of you tests so here's some examples of you
  tests so here's some examples of you
know not good better and best um it's know not good better and best um it's
  know not good better and best um it's
actually so this is a weird anapat like actually so this is a weird anapat like
  actually so this is a weird anapat like
you shouldn't you probably shouldn't be you shouldn't you probably shouldn't be
  you shouldn't you probably shouldn't be
using Global State at all like making a using Global State at all like making a
  using Global State at all like making a
globally modifiable variable seems worse globally modifiable variable seems worse
  globally modifiable variable seems worse
than making a constant um the reason than making a constant um the reason
  than making a constant um the reason
it's usually better is because constants it's usually better is because constants
  it's usually better is because constants
are they really limit your testability are they really limit your testability
  are they really limit your testability
and so making it a variable at least you and so making it a variable at least you
  and so making it a variable at least you
could change some behavior in a test um could change some behavior in a test um
  could change some behavior in a test um
but the best is actually making the top but the best is actually making the top
  but the best is actually making the top
level a constant and then making some level a constant and then making some
  level a constant and then making some
sort of way to modify that uh in another sort of way to modify that uh in another
  sort of way to modify that uh in another
way okay test helpers um so here's an way okay test helpers um so here's an
  way okay test helpers um so here's an
example of a test helper that creates a example of a test helper that creates a
  example of a test helper that creates a
temporary file um so there's a few temporary file um so there's a few
  temporary file um so there's a few
properties of this test helper that um properties of this test helper that um
  properties of this test helper that um
I'll talk about so one the test help I'll talk about so one the test help
  I'll talk about so one the test help
test helpers should never return an test helpers should never return an
  test helpers should never return an
error um you know functions and go error um you know functions and go
  error um you know functions and go
should always return errors test helpers should always return errors test helpers
  should always return errors test helpers
should never ever return an error they should never ever return an error they
  should never ever return an error they
have X they should have access to the T have X they should have access to the T
  have X they should have access to the T
structure so just fail the test if structure so just fail the test if
  structure so just fail the test if
there's an error um the other thing is there's an error um the other thing is
  there's an error um the other thing is
uh sort of by not returning the error uh sort of by not returning the error
  uh sort of by not returning the error
you make the test usage a lot clearer if you make the test usage a lot clearer if
  you make the test usage a lot clearer if
if your test helpers return an error if your test helpers return an error
  if your test helpers return an error
your tests that use the helpers turn your tests that use the helpers turn
  your tests that use the helpers turn
into run test helper if air not equal n into run test helper if air not equal n
  into run test helper if air not equal n
you know fatal over and over and over you know fatal over and over and over
  you know fatal over and over and over
whereas when they just fail on their own whereas when they just fail on their own
  whereas when they just fail on their own
you could get a really dense block of you could get a really dense block of
  you could get a really dense block of
test helper calls you know right at the test helper calls you know right at the
  test helper calls you know right at the
beginning to do a bunch of setup and beginning to do a bunch of setup and
  beginning to do a bunch of setup and
it's just a lot it removes a lot of I it's just a lot it removes a lot of I
  it's just a lot it removes a lot of I
would say like visual over head when would say like visual over head when
  would say like visual over head when
you're trying to figure out what a test you're trying to figure out what a test
  you're trying to figure out what a test
does um and you should be careful about does um and you should be careful about
  does um and you should be careful about
when you make these test helpers I guess when you make these test helpers I guess
  when you make these test helpers I guess
I get to that um so I'll talk about that I get to that um so I'll talk about that
  I get to that um so I'll talk about that
later the other thing I want to point later the other thing I want to point
  later the other thing I want to point
out is I use a feature that's in go 1.9 out is I use a feature that's in go 1.9
  out is I use a feature that's in go 1.9
um t. helper uh so that just just was um t. helper uh so that just just was
  um t. helper uh so that just just was
introduced in 1.9 but by calling t. introduced in 1.9 but by calling t.
  introduced in 1.9 but by calling t.
helper in in a test helper it makes the helper in in a test helper it makes the
  helper in in a test helper it makes the
error out the stack Trace output better error out the stack Trace output better
  error out the stack Trace output better
um when some you know panic situation um when some you know panic situation
  um when some you know panic situation
happens on a test helper uh test helpers happens on a test helper uh test helpers
  happens on a test helper uh test helpers
have been notorious in the past for have been notorious in the past for
  have been notorious in the past for
panicking and the failures there but panicking and the failures there but
  panicking and the failures there but
it's sometimes hard to find the exact it's sometimes hard to find the exact
  it's sometimes hard to find the exact
test where that failed so if you're test where that failed so if you're
  test where that failed so if you're
moving on to go 1.9 I recommend just moving on to go 1.9 I recommend just
  moving on to go 1.9 I recommend just
dropping this in all your test dropping this in all your test
  dropping this in all your test
helpers uh and so I do I explained all helpers uh and so I do I explained all
  helpers uh and so I do I explained all
this stuff so I'm going to skip it um this stuff so I'm going to skip it um
  this stuff so I'm going to skip it um
the other neat trick is if you have the other neat trick is if you have
  the other neat trick is if you have
cleanup to do you could return a closure cleanup to do you could return a closure
  cleanup to do you could return a closure
and we usually just return a funk like and we usually just return a funk like
  and we usually just return a funk like
no return value closure to actually do no return value closure to actually do
  no return value closure to actually do
the cleanup so in this example here's the cleanup so in this example here's
  the cleanup so in this example here's
the temp file where as cleanup we the temp file where as cleanup we
  the temp file where as cleanup we
actually want to delete the temp file actually want to delete the temp file
  actually want to delete the temp file
and so what we do is we do all the setup and so what we do is we do all the setup
  and so what we do is we do all the setup
we airor if we have to and what we do is we airor if we have to and what we do is
  we airor if we have to and what we do is
return the temp file we return a closure return the temp file we return a closure
  return the temp file we return a closure
that deletes it um and the benefit of that deletes it um and the benefit of
  that deletes it um and the benefit of
this is that closure since it is a this is that closure since it is a
  this is that closure since it is a
closure has access still to that t um T closure has access still to that t um T
  closure has access still to that t um T
value so it could check the error like value so it could check the error like
  value so it could check the error like
in in the production code where we use in in the production code where we use
  in in the production code where we use
this we check the error of os remove um this we check the error of os remove um
  this we check the error of os remove um
OS remove can potentially fail and if it OS remove can potentially fail and if it
  OS remove can potentially fail and if it
fails we do a t. fatal F and you don't fails we do a t. fatal F and you don't
  fails we do a t. fatal F and you don't
have to pass that back cuz we we already have to pass that back cuz we we already
  have to pass that back cuz we we already
closed over it and so when you use it it closed over it and so when you use it it
  closed over it and so when you use it it
ends up looking uh like this um ends up looking uh like this um
  ends up looking uh like this um
sometimes you get really clever um and sometimes you get really clever um and
  sometimes you get really clever um and
it's not always beneficial for it's not always beneficial for
  it's not always beneficial for
readability but sometimes you can get readability but sometimes you can get
  readability but sometimes you can get
really clever uh if you're setting up really clever uh if you're setting up
  really clever uh if you're setting up
some side effect some some World state some side effect some some World state
  some side effect some some World state
that actually has no return value um of that actually has no return value um of
  that actually has no return value um of
just returning the function alone you just returning the function alone you
  just returning the function alone you
could oneline the whole thing and so could oneline the whole thing and so
  could oneline the whole thing and so
this is a more or less I mean I think this is a more or less I mean I think
  this is a more or less I mean I think
it's a little bit longer our actual one it's a little bit longer our actual one
  it's a little bit longer our actual one
because we check all the errors but because we check all the errors but
  because we check all the errors but
there's more or less a test seler we there's more or less a test seler we
  there's more or less a test seler we
have in in all our projects uh where we have in in all our projects uh where we
  have in in all our projects uh where we
could change a directory uh to test could change a directory uh to test
  could change a directory uh to test
things and and usually we're changing things and and usually we're changing
  things and and usually we're changing
directories to test CLI behavior um but directories to test CLI behavior um but
  directories to test CLI behavior um but
you could oneline it the negative aspect you could oneline it the negative aspect
  you could oneline it the negative aspect
of this is a new developer coming into of this is a new developer coming into
  of this is a new developer coming into
the project looking at that line it's the project looking at that line it's
  the project looking at that line it's
not super obvious like that what that's not super obvious like that what that's
  not super obvious like that what that's
doing um so that's a downside the the doing um so that's a downside the the
  doing um so that's a downside the the
argument we sometimes make is that we we argument we sometimes make is that we we
  argument we sometimes make is that we we
really only use it I think for CH dur uh really only use it I think for CH dur uh
  really only use it I think for CH dur uh
and it's we think this particular cases and it's we think this particular cases
  and it's we think this particular cases
obvious obvious
  obvious
enough so I said all this um most enough so I said all this um most
  enough so I said all this um most
important thing is you have use the important thing is you have use the
  important thing is you have use the
testing. T just use it in your test testing. T just use it in your test
  testing. T just use it in your test
helpers uh don't return errors in your helpers uh don't return errors in your
  helpers uh don't return errors in your
test test
  test
helpers okay this is uh I guess I should helpers okay this is uh I guess I should
  helpers okay this is uh I guess I should
also say that some of these I expect also say that some of these I expect
  also say that some of these I expect
people will disagree with so this is one people will disagree with so this is one
  people will disagree with so this is one
of those that that I think multiple of those that that I think multiple
  of those that that I think multiple
times at hhic Corp we've hired very times at hhic Corp we've hired very
  times at hhic Corp we've hired very
experienced go engineers and this is experienced go engineers and this is
  experienced go engineers and this is
something that have has sometimes uh something that have has sometimes uh
  something that have has sometimes uh
rubbed them the wrong way but it makes rubbed them the wrong way but it makes
  rubbed them the wrong way but it makes
over time we found everyone kind of over time we found everyone kind of
  over time we found everyone kind of
likes us more so repeat yourself in likes us more so repeat yourself in
  likes us more so repeat yourself in
tests um what what we prefer what I tests um what what we prefer what I
  tests um what what we prefer what I
prefer overall in test is localize logic prefer overall in test is localize logic
  prefer overall in test is localize logic
as much as possible when a test fails I as much as possible when a test fails I
  as much as possible when a test fails I
usually or I or someone else usually usually or I or someone else usually
  usually or I or someone else usually
wrote that test a long time ago um and wrote that test a long time ago um and
  wrote that test a long time ago um and
it's failing because I'm doing a it's failing because I'm doing a
  it's failing because I'm doing a
refactor I'm doing a new feature that refactor I'm doing a new feature that
  refactor I'm doing a new feature that
caused it to fail and I don't really caused it to fail and I don't really
  caused it to fail and I don't really
remember the details of the tests remember the details of the tests
  remember the details of the tests
anymore and there's nothing more anymore and there's nothing more
  anymore and there's nothing more
frustrating to me than going to a test frustrating to me than going to a test
  frustrating to me than going to a test
test and realizing it calls seven test and realizing it calls seven
  test and realizing it calls seven
different functions that are in four different functions that are in four
  different functions that are in four
different files and I have to like start different files and I have to like start
  different files and I have to like start
building all this mental context of building all this mental context of
  building all this mental context of
what's Happening why is it doing that what's Happening why is it doing that
  what's Happening why is it doing that
that sort of thing uh it's much easier that sort of thing uh it's much easier
  that sort of thing uh it's much easier
actually uh I think to just have a huge actually uh I think to just have a huge
  actually uh I think to just have a huge
test that's like 200 300 lines of code test that's like 200 300 lines of code
  test that's like 200 300 lines of code
that just does everything right there so that just does everything right there so
  that just does everything right there so
that I could just go through it and that I could just go through it and
  that I could just go through it and
understand exactly what that test is um understand exactly what that test is um
  understand exactly what that test is um
the project where we do this the most is the project where we do this the most is
  the project where we do this the most is
terraform uh if you go in terraform terraform uh if you go in terraform
  terraform uh if you go in terraform
there's files called context underscore there's files called context underscore
  there's files called context underscore
any of the tests in there uh the each any of the tests in there uh the each
  any of the tests in there uh the each
test is around 200 lines of code and test is around 200 lines of code and
  test is around 200 lines of code and
when we write a new test what we do is when we write a new test what we do is
  when we write a new test what we do is
we take an old test we copy it we paste we take an old test we copy it we paste
  we take an old test we copy it we paste
it and then we start modifying the five it and then we start modifying the five
  it and then we start modifying the five
lines we need to modify um and it feels lines we need to modify um and it feels
  lines we need to modify um and it feels
bad sometimes but it's been you know bad sometimes but it's been you know
  bad sometimes but it's been you know
four years and we get a lot of test four years and we get a lot of test
  four years and we get a lot of test
failures in context because it's sort of failures in context because it's sort of
  failures in context because it's sort of
the core of terraform and it's super the core of terraform and it's super
  the core of terraform and it's super
super helpful to be able to open one you super helpful to be able to open one you
  super helpful to be able to open one you
know Vim panel look and have everything know Vim panel look and have everything
  know Vim panel look and have everything
there that you need to figure out why there that you need to figure out why
  there that you need to figure out why
the test is the test is
  the test is
failing failing
  failing
so copy and paste for tests um we don't so copy and paste for tests um we don't
  so copy and paste for tests um we don't
we don't practice repeat yourself too we don't practice repeat yourself too
  we don't practice repeat yourself too
much in actual non- test code but in much in actual non- test code but in
  much in actual non- test code but in
tests we do prefer uh 200 line test to a tests we do prefer uh 200 line test to a
  tests we do prefer uh 200 line test to a
20 line test with abstracted 20 line test with abstracted
  20 line test with abstracted
helpers packages and functions um I saw helpers packages and functions um I saw
  helpers packages and functions um I saw
I saw on Twitter some slides about anti I saw on Twitter some slides about anti
  I saw on Twitter some slides about anti
patterns that talked about some of this patterns that talked about some of this
  patterns that talked about some of this
and I for tests I agree with some of and I for tests I agree with some of
  and I for tests I agree with some of
them and then some other ones uh it's them and then some other ones uh it's
  them and then some other ones uh it's
it's questionable so um you know break it's questionable so um you know break
  it's questionable so um you know break
this is hard because it's going to be this is hard because it's going to be
  this is hard because it's going to be
you get this with experience with go you you get this with experience with go you
  you get this with experience with go you
can't there's no real hard and fast can't there's no real hard and fast
  can't there's no real hard and fast
rules you could come in to go and know rules you could come in to go and know
  rules you could come in to go and know
exactly when to make a package and when exactly when to make a package and when
  exactly when to make a package and when
not to um and so what you want to be not to um and so what you want to be
  not to um and so what you want to be
able to do is break down functionality able to do is break down functionality
  able to do is break down functionality
into packages and functions when it into packages and functions when it
  into packages and functions when it
makes sense the when it makes sense is makes sense the when it makes sense is
  makes sense the when it makes sense is
super difficult to explain um you also super difficult to explain um you also
  super difficult to explain um you also
don't want to overdo it um which is don't want to overdo it um which is
  don't want to overdo it um which is
super hard to explain um but doing this super hard to explain um but doing this
  super hard to explain um but doing this
correctly AIDS testing correctly AIDS testing
  correctly AIDS testing
quite a bit because it gives you a much quite a bit because it gives you a much
  quite a bit because it gives you a much
cleaner surface area of what to test um cleaner surface area of what to test um
  cleaner surface area of what to test um
you know you know there's certain safety you know you know there's certain safety
  you know you know there's certain safety
mechanisms of package boundaries and mechanisms of package boundaries and
  mechanisms of package boundaries and
unexported versus exported that you unexported versus exported that you
  unexported versus exported that you
don't need to worry about um and and it don't need to worry about um and and it
  don't need to worry about um and and it
makes things a lot easier so a good makes things a lot easier so a good
  makes things a lot easier so a good
example here is terraform has a dag example here is terraform has a dag
  example here is terraform has a dag
package for doing uh directed ayylic package for doing uh directed ayylic
  package for doing uh directed ayylic
graphs and that dag package was written graphs and that dag package was written
  graphs and that dag package was written
I don't know in 2013 or 2012 uh and it's I don't know in 2013 or 2012 uh and it's
  I don't know in 2013 or 2012 uh and it's
been it's been touched like three times been it's been touched like three times
  been it's been touched like three times
in four years because because writing in four years because because writing
  in four years because because writing
writing a simple like graph Library writing a simple like graph Library
  writing a simple like graph Library
doesn't usually require change um so we doesn't usually require change um so we
  doesn't usually require change um so we
write the test there we know the write the test there we know the
  write the test there we know the
coverage of that package is 100% at all coverage of that package is 100% at all
  coverage of that package is 100% at all
times um and we know the bugs likely times um and we know the bugs likely
  times um and we know the bugs likely
aren't there um but it's it's there aren't there um but it's it's there
  aren't there um but it's it's there
there's some issues with this like there's some issues with this like
  there's some issues with this like
package ification so uh oh I actually package ification so uh oh I actually
  package ification so uh oh I actually
didn't bring this in here it's the next didn't bring this in here it's the next
  didn't bring this in here it's the next
slide um but unless the the function is slide um but unless the the function is
  slide um but unless the the function is
extremely complex um we usually try to extremely complex um we usually try to
  extremely complex um we usually try to
only test exported functions um we view only test exported functions um we view
  only test exported functions um we view
sort of the exported API as the place sort of the exported API as the place
  sort of the exported API as the place
where we need to do the tests but for where we need to do the tests but for
  where we need to do the tests but for
internal helpers that are that are internal helpers that are that are
  internal helpers that are that are
really complicated or fan out quite a really complicated or fan out quite a
  really complicated or fan out quite a
bit or have a lot of edge cases we will bit or have a lot of edge cases we will
  bit or have a lot of edge cases we will
test those internal functions um some test those internal functions um some
  test those internal functions um some
people do take this too far uh the way people do take this too far uh the way
  people do take this too far uh the way
to take this too far is that the that to take this too far is that the that
  to take this too far is that the that
you only do like sort of integration or you only do like sort of integration or
  you only do like sort of integration or
acceptance test level it's like if I acceptance test level it's like if I
  acceptance test level it's like if I
just test the blackbox behavior then I just test the blackbox behavior then I
  just test the blackbox behavior then I
know it works um I personally believe know it works um I personally believe
  know it works um I personally believe
that's a little bit too far so we try to that's a little bit too far so we try to
  that's a little bit too far so we try to
B we do acceptance test as well so we B we do acceptance test as well so we
  B we do acceptance test as well so we
try to balance out the acceptance test try to balance out the acceptance test
  try to balance out the acceptance test
usage with testing the exported API with usage with testing the exported API with
  usage with testing the exported API with
uh um with sort of um educated you know uh um with sort of um educated you know
  uh um with sort of um educated you know
choices of testing internal apis as choices of testing internal apis as
  choices of testing internal apis as
well um yeah we usually treat the well um yeah we usually treat the
  well um yeah we usually treat the
unexported sort of functions and structs unexported sort of functions and structs
  unexported sort of functions and structs
as implementation details um and by not as implementation details um and by not
  as implementation details um and by not
testing those specifically it makes testing those specifically it makes
  testing those specifically it makes
refactoring a lot easier in the future refactoring a lot easier in the future
  refactoring a lot easier in the future
but you want their logic like what the but you want their logic like what the
  but you want their logic like what the
what they're trying to achieve their what they're trying to achieve their
  what they're trying to achieve their
goals they're trying to achieve tested goals they're trying to achieve tested
  goals they're trying to achieve tested
via the exported API and if you can't do via the exported API and if you can't do
  via the exported API and if you can't do
that you probably have to unit test the that you probably have to unit test the
  that you probably have to unit test the
internal internal
  internal
ones um so there's internal packages and ones um so there's internal packages and
  ones um so there's internal packages and
I don't remember what Go version this I don't remember what Go version this
  I don't remember what Go version this
was introduced it's like 1 was introduced it's like 1
  was introduced it's like 1
145 um but if you make a package named 145 um but if you make a package named
  145 um but if you make a package named
internal then any pack I can't explain internal then any pack I can't explain
  internal then any pack I can't explain
it in one sentence probably but any any it in one sentence probably but any any
  it in one sentence probably but any any
packages sort of under that can only be packages sort of under that can only be
  packages sort of under that can only be
accessed by packages in folders less accessed by packages in folders less
  accessed by packages in folders less
than the folder contain it doesn't than the folder contain it doesn't
  than the folder contain it doesn't
matter so there's there's there there's matter so there's there's there there's
  matter so there's there's there there's
internal packages they allow you to to internal packages they allow you to to
  internal packages they allow you to to
create a mechanism whereby external create a mechanism whereby external
  create a mechanism whereby external
people cannot import your internal people cannot import your internal
  people cannot import your internal
packages um and we like to use this packages um and we like to use this
  packages um and we like to use this
actually as a mechanism through to actually as a mechanism through to
  actually as a mechanism through to
uncomfortably over package things um the uncomfortably over package things um the
  uncomfortably over package things um the
and so this is where um I saw a slide and so this is where um I saw a slide
  and so this is where um I saw a slide
yesterday that said don't over package yesterday that said don't over package
  yesterday that said don't over package
if anything under package um and we used if anything under package um and we used
  if anything under package um and we used
to do that and the issue we ran into to do that and the issue we ran into
  to do that and the issue we ran into
terraform as a living embodiment of that terraform as a living embodiment of that
  terraform as a living embodiment of that
issue right now is terraform under issue right now is terraform under
  issue right now is terraform under
packaged and it's basically impossible packaged and it's basically impossible
  packaged and it's basically impossible
right now for us to re-lit out all the right now for us to re-lit out all the
  right now for us to re-lit out all the
packages we see in terraform without packages we see in terraform without
  packages we see in terraform without
doing it atomically because every time doing it atomically because every time
  doing it atomically because every time
we try to do it incrementally we just we try to do it incrementally we just
  we try to do it incrementally we just
get import Cycles because everything is get import Cycles because everything is
  get import Cycles because everything is
tying into everything um and so if you tying into everything um and so if you
  tying into everything um and so if you
look at Vault or something in our Ford look at Vault or something in our Ford
  look at Vault or something in our Ford
things we over package from the things we over package from the
  things we over package from the
beginning and we ended up deleting some beginning and we ended up deleting some
  beginning and we ended up deleting some
packages because they didn't make sense packages because they didn't make sense
  packages because they didn't make sense
as packages but we have a much better as packages but we have a much better
  as packages but we have a much better
package breakdown um and one of the ways package breakdown um and one of the ways
  package breakdown um and one of the ways
to do this safely without creating to do this safely without creating
  to do this safely without creating
public API for for you to promise people public API for for you to promise people
  public API for for you to promise people
anything is to use internal packages anything is to use internal packages
  anything is to use internal packages
that way you know if you have I'm that way you know if you have I'm
  that way you know if you have I'm
exaggerating here but if you have like exaggerating here but if you have like
  exaggerating here but if you have like
500 internal packages to the end user 500 internal packages to the end user
  500 internal packages to the end user
still looks like one um so I would say still looks like one um so I would say
  still looks like one um so I would say
you know do again it's the package you know do again it's the package
  you know do again it's the package
boundary thing is so qualitative um and boundary thing is so qualitative um and
  boundary thing is so qualitative um and
so experience-based that it's hard for so experience-based that it's hard for
  so experience-based that it's hard for
me to give any hard and fast rules um me to give any hard and fast rules um
  me to give any hard and fast rules um
but the package boundary stuff will help but the package boundary stuff will help
  but the package boundary stuff will help
testing quite a testing quite a
  testing quite a
bit um so networking how do we test bit um so networking how do we test
  bit um so networking how do we test
networking we have all our software I networking we have all our software I
  networking we have all our software I
think actually does networking to some think actually does networking to some
  think actually does networking to some
extent and so what we want to be able to extent and so what we want to be able to
  extent and so what we want to be able to
do is test network connections and so if do is test network connections and so if
  do is test network connections and so if
you're testing networking make a real you're testing networking make a real
  you're testing networking make a real
network connection there's I don't see network connection there's I don't see
  network connection there's I don't see
it actually very often anymore so I it actually very often anymore so I
  it actually very often anymore so I
think we're we're in a really good spot think we're we're in a really good spot
  think we're we're in a really good spot
but early on at least in go code in 2012 but early on at least in go code in 2012
  but early on at least in go code in 2012
and so on I saw a lot of people trying and so on I saw a lot of people trying
  and so on I saw a lot of people trying
to mock net.com net.com is an interface to mock net.com net.com is an interface
  to mock net.com net.com is an interface
that seems like a perfect place to that seems like a perfect place to
  that seems like a perfect place to
create a mock um but there's really no create a mock um but there's really no
  create a mock um but there's really no
point uh to mocking net.com and I'll point uh to mocking net.com and I'll
  point uh to mocking net.com and I'll
explain so here's a fully functional I explain so here's a fully functional I
  explain so here's a fully functional I
think yeah I think it works uh think yeah I think it works uh
  think yeah I think it works uh
connection helper to create a test connection helper to create a test
  connection helper to create a test
connection with the client and server connection with the client and server
  connection with the client and server
end and all we do is listen we use the end and all we do is listen we use the
  end and all we do is listen we use the
operating system to choose our port for operating system to choose our port for
  operating system to choose our port for
us um we bind it to Local Host so we us um we bind it to Local Host so we
  us um we bind it to Local Host so we
only we can connect ourselves but it's only we can connect ourselves but it's
  only we can connect ourselves but it's
you know just choose a port I don't care you know just choose a port I don't care
  you know just choose a port I don't care
you immediately make the connection um you immediately make the connection um
  you immediately make the connection um
the accept you can see is not in a for the accept you can see is not in a for
  the accept you can see is not in a for
Loop so the moment that we accept the Loop so the moment that we accept the
  Loop so the moment that we accept the
connection we close the listener which connection we close the listener which
  connection we close the listener which
does not close the connections we just does not close the connections we just
  does not close the connections we just
don't accept anymore um and then we dial don't accept anymore um and then we dial
  don't accept anymore um and then we dial
it at the end and we just return the it at the end and we just return the
  it at the end and we just return the
client server and you go close them on client server and you go close them on
  client server and you go close them on
the reverse side um and so that was a the reverse side um and so that was a
  the reverse side um and so that was a
that was a one connection example it's that was a one connection example it's
  that was a one connection example it's
super easy to make an End connection super easy to make an End connection
  super easy to make an End connection
test helper um it's easy to test any test helper um it's easy to test any
  test helper um it's easy to test any
protocol this way we actually in Packer protocol this way we actually in Packer
  protocol this way we actually in Packer
for example we have this to test SSH for example we have this to test SSH
  for example we have this to test SSH
connections and the way we test s SSH connections and the way we test s SSH
  connections and the way we test s SSH
connections is we create a real SSH connections is we create a real SSH
  connections is we create a real SSH
server we create a listener we connect server we create a listener we connect
  server we create a listener we connect
to it we shut down the SSH so we return to it we shut down the SSH so we return
  to it we shut down the SSH so we return
the two connections and so you have a the two connections and so you have a
  the two connections and so you have a
real SSH connection um it's easy to real SSH connection um it's easy to
  real SSH connection um it's easy to
return turn the listener if you need the return turn the listener if you need the
  return turn the listener if you need the
listener it's easy to test multiple listener it's easy to test multiple
  listener it's easy to test multiple
networking types IPv6 or Unix domain networking types IPv6 or Unix domain
  networking types IPv6 or Unix domain
sockets or other things it's all there sockets or other things it's all there
  sockets or other things it's all there
um and then rhetorically at the end you um and then rhetorically at the end you
  um and then rhetorically at the end you
know it's sort of like there's no reason know it's sort of like there's no reason
  know it's sort of like there's no reason
to ever mock a 
  
netcon uh configurability on the how to netcon uh configurability on the how to
  netcon uh configurability on the how to
write testable code side um so write testable code side um so
  write testable code side um so
unconfigured behavior is very often the unconfigured behavior is very often the
  unconfigured behavior is very often the
point of difficulty for tests uh your point of difficulty for tests uh your
  point of difficulty for tests uh your
code is doing something um that may code is doing something um that may
  code is doing something um that may
totally make sense for produ usage uh so totally make sense for produ usage uh so
  totally make sense for produ usage uh so
it's not configurable but the tests want it's not configurable but the tests want
  it's not configurable but the tests want
to do change the behavior make it faster to do change the behavior make it faster
  to do change the behavior make it faster
skip some safety test I don't know um skip some safety test I don't know um
  skip some safety test I don't know um
and so what we usually do is overp and so what we usually do is overp
  and so what we usually do is overp
parameterize our strs to allow test to parameterize our strs to allow test to
  parameterize our strs to allow test to
fine-tune behavior um and if you really fine-tune behavior um and if you really
  fine-tune behavior um and if you really
don't want users to do this just make don't want users to do this just make
  don't want users to do this just make
the parameters on the strs unexported um the parameters on the strs unexported um
  the parameters on the strs unexported um
another thing we do and I don't know another thing we do and I don't know
  another thing we do and I don't know
yeah I don't think I showed in the yeah I don't think I showed in the
  yeah I don't think I showed in the
example um is we prefix sometimes these example um is we prefix sometimes these
  example um is we prefix sometimes these
with test so that people know that their with test so that people know that their
  with test so that people know that their
parameters only to be used with test parameters only to be used with test
  parameters only to be used with test
um but by making them over um but by making them over
  um but by making them over
overparameterized in the beginning it overparameterized in the beginning it
  overparameterized in the beginning it
just makes that a lot just makes that a lot
  just makes that a lot
easier um so here's here's an example easier um so here's here's an example
  easier um so here's here's an example
like cash path and port in this example like cash path and port in this example
  like cash path and port in this example
they may always be the same they may they may always be the same they may
  they may always be the same they may
never in production they may maybe never in production they may maybe
  never in production they may maybe
should never change um you should still should never change um you should still
  should never change um you should still
make them configurable because in make them configurable because in
  make them configurable because in
testing you probably do want to change 
  
them oh I don't remember what this them oh I don't remember what this
  them oh I don't remember what this
is no I'm going to skip that I don't is no I'm going to skip that I don't
  is no I'm going to skip that I don't
remember what that's there for um so remember what that's there for um so
  remember what that's there for um so
another thing we do pretty commonly is another thing we do pretty commonly is
  another thing we do pretty commonly is
just make a big like test Bull and the just make a big like test Bull and the
  just make a big like test Bull and the
Comon above it should very should very Comon above it should very should very
  Comon above it should very should very
clearly Define what Behavior that's clearly Define what Behavior that's
  clearly Define what Behavior that's
going to change um but the most common going to change um but the most common
  going to change um but the most common
way uh I've seen this used uh we've used way uh I've seen this used uh we've used
  way uh I've seen this used uh we've used
it a couple times but the most common it a couple times but the most common
  it a couple times but the most common
way I've seen this used is to like skip way I've seen this used is to like skip
  way I've seen this used is to like skip
off in in a web application or something off in in a web application or something
  off in in a web application or something
like say your web application's only like say your web application's only
  like say your web application's only
mechanism for logging in is ooth um mechanism for logging in is ooth um
  mechanism for logging in is ooth um
testing or simulating ooth completely is testing or simulating ooth completely is
  testing or simulating ooth completely is
kind of a bear and so instead of doing kind of a bear and so instead of doing
  kind of a bear and so instead of doing
that we could just set test equals true that we could just set test equals true
  that we could just set test equals true
and that'll always ooth you as the same and that'll always ooth you as the same
  and that'll always ooth you as the same
person you know ooth it doesn't actually person you know ooth it doesn't actually
  person you know ooth it doesn't actually
do any ooth protocol stuff it just gets do any ooth protocol stuff it just gets
  do any ooth protocol stuff it just gets
the request to ooth and immediately the request to ooth and immediately
  the request to ooth and immediately
pretends you're logged in right 
  
away right complex structs testing away right complex structs testing
  away right complex structs testing
complex structs uh struct values so uh complex structs uh struct values so uh
  complex structs uh struct values so uh
yeah so an example here would be yeah so an example here would be
  yeah so an example here would be
terraform has a graph structure um and terraform has a graph structure um and
  terraform has a graph structure um and
the graph structure has a lot of the graph structure has a lot of
  the graph structure has a lot of
obviously you know pointers to Children obviously you know pointers to Children
  obviously you know pointers to Children
there's data on the nodes themselves and there's data on the nodes themselves and
  there's data on the nodes themselves and
we want to be able to test that one we want to be able to test that one
  we want to be able to test that one
graph equals another graph um the blunt graph equals another graph um the blunt
  graph equals another graph um the blunt
instrument that everyone use we use a instrument that everyone use we use a
  instrument that everyone use we use a
lot um to test complex structs is lot um to test complex structs is
  lot um to test complex structs is
reflect. deal um you just use that and reflect. deal um you just use that and
  reflect. deal um you just use that and
verify the structure the same um a verify the structure the same um a
  verify the structure the same um a
slightly better approach uh is to is slightly better approach uh is to is
  slightly better approach uh is to is
there's actually a number of really good there's actually a number of really good
  there's actually a number of really good
thirdparty libraries now that do stru thirdparty libraries now that do stru
  thirdparty libraries now that do stru
comparison and generate nicer error comparison and generate nicer error
  comparison and generate nicer error
messages when they don't equal I'm sure messages when they don't equal I'm sure
  messages when they don't equal I'm sure
a lot of people here have run into the a lot of people here have run into the
  a lot of people here have run into the
like hour hopefully an hour or less but like hour hopefully an hour or less but
  like hour hopefully an hour or less but
maybe more our loss When reflect. de maybe more our loss When reflect. de
  maybe more our loss When reflect. de
deep equal returns false and it's deep equal returns false and it's
  deep equal returns false and it's
because like you had an INT and then the because like you had an INT and then the
  because like you had an INT and then the
actual type was like in 64 um and that actual type was like in 64 um and that
  actual type was like in 64 um and that
causes that to fail and it causes you to causes that to fail and it causes you to
  causes that to fail and it causes you to
rip your hair out because you're dumping rip your hair out because you're dumping
  rip your hair out because you're dumping
they're dumping the structs and you're they're dumping the structs and you're
  they're dumping the structs and you're
looking bite by bite and they're like looking bite by bite and they're like
  looking bite by bite and they're like
they're the same they're the same struct they're the same they're the same struct
  they're the same they're the same struct
um and stuff like that so there's better um and stuff like that so there's better
  um and stuff like that so there's better
ways to do it um one thing we ways to do it um one thing we
  ways to do it um one thing we
do um at at times not blanket across all do um at at times not blanket across all
  do um at at times not blanket across all
complex structs where it makes sense is complex structs where it makes sense is
  complex structs where it makes sense is
we use this pattern of test string it's we use this pattern of test string it's
  we use this pattern of test string it's
like the ghost Stringer interface but like the ghost Stringer interface but
  like the ghost Stringer interface but
for tests and unexported so we do this for tests and unexported so we do this
  for tests and unexported so we do this
lowercase test string thing you use btes lowercase test string thing you use btes
  lowercase test string thing you use btes
buffer fumed or anything to create some buffer fumed or anything to create some
  buffer fumed or anything to create some
humanfriendly output that's testing the humanfriendly output that's testing the
  humanfriendly output that's testing the
things that matter uh and then when things that matter uh and then when
  things that matter uh and then when
you're testing that complex thing you you're testing that complex thing you
  you're testing that complex thing you
actually just compare the strings to actually just compare the strings to
  actually just compare the strings to
each other uh and a good example of each other uh and a good example of
  each other uh and a good example of
where we use this is actually the terar where we use this is actually the terar
  where we use this is actually the terar
from graph from graph
  from graph
so the graph route or any graph node has so the graph route or any graph node has
  so the graph route or any graph node has
this test string on it when you call this test string on it when you call
  this test string on it when you call
test string we generate a human readable test string we generate a human readable
  test string we generate a human readable
sort of like indented structure to sort of like indented structure to
  sort of like indented structure to
represent the graph um because that's represent the graph um because that's
  represent the graph um because that's
what we care about the most uh when what we care about the most uh when
  what we care about the most uh when
we're testing those things and that's a we're testing those things and that's a
  we're testing those things and that's a
lot I mean reflect deep equal would lot I mean reflect deep equal would
  lot I mean reflect deep equal would
check too many fields that we don't care check too many fields that we don't care
  check too many fields that we don't care
if they match um and you get into a lot if they match um and you get into a lot
  if they match um and you get into a lot
more complex more complex
  more complex
things so for data structures like trees things so for data structures like trees
  things so for data structures like trees
link lists Etc um this sort of pattern link lists Etc um this sort of pattern
  link lists Etc um this sort of pattern
helps helps a lot uh and like I said you helps helps a lot uh and like I said you
  helps helps a lot uh and like I said you
could use reflect deepical or third could use reflect deepical or third
  could use reflect deepical or third
party lib we certainly do that in party lib we certainly do that in
  party lib we certainly do that in
different places um and going to be different places um and going to be
  different places um and going to be
honest that the test string thing is honest that the test string thing is
  honest that the test string thing is
pretty blunt like you when people see it pretty blunt like you when people see it
  pretty blunt like you when people see it
it's kind of weird um but we've had it's kind of weird um but we've had
  it's kind of weird um but we've had
really good results for complex really good results for complex
  really good results for complex
trucks and here's an example of like the trucks and here's an example of like the
  trucks and here's an example of like the
test string output we generate for our test string output we generate for our
  test string output we generate for our
graphs in terraform it's like a a simple graphs in terraform it's like a a simple
  graphs in terraform it's like a a simple
graph testing a simple single dependency graph testing a simple single dependency
  graph testing a simple single dependency
uh and these These are the strings We uh and these These are the strings We
  uh and these These are the strings We
compare and so when they fail it's compare and so when they fail it's
  compare and so when they fail it's
actually a lot easier for us to generate actually a lot easier for us to generate
  actually a lot easier for us to generate
diffs and failure output that helps diffs and failure output that helps
  diffs and failure output that helps
versus you know these two graph versus you know these two graph
  versus you know these two graph
structures that together have I don't structures that together have I don't
  structures that together have I don't
know well this is a simple one but in in know well this is a simple one but in in
  know well this is a simple one but in in
the more complex terraform tests like the more complex terraform tests like
  the more complex terraform tests like
for example a graph could very easily for example a graph could very easily
  for example a graph could very easily
have 2,000 or more graph nodes in there have 2,000 or more graph nodes in there
  have 2,000 or more graph nodes in there
and a reflect dpal failure is a and a reflect dpal failure is a
  and a reflect dpal failure is a
nightmare to a bug whereas this is a lot 
  
easier subprocessing um this one is easier subprocessing um this one is
  easier subprocessing um this one is
another one I specifically remember I I another one I specifically remember I I
  another one I specifically remember I I
got from the goh standard Library got from the goh standard Library
  got from the goh standard Library
Library um and thought was genius Library um and thought was genius
  Library um and thought was genius
whoever did that in the ghost standard whoever did that in the ghost standard
  whoever did that in the ghost standard
Library so subprocessing is a typical Library so subprocessing is a typical
  Library so subprocessing is a typical
point of difficult to test Behavior very point of difficult to test Behavior very
  point of difficult to test Behavior very
very typical um and there's usually two very typical um and there's usually two
  very typical um and there's usually two
options when when you're faced with the options when when you're faced with the
  options when when you're faced with the
need to subprocess you could either need to subprocess you could either
  need to subprocess you could either
actually do the subprocess or you could actually do the subprocess or you could
  actually do the subprocess or you could
mock the output or behavior or mock the mock the output or behavior or mock the
  mock the output or behavior or mock the
whole subprocess in general so um as an whole subprocess in general so um as an
  whole subprocess in general so um as an
example like say you're writing an example like say you're writing an
  example like say you're writing an
application that interfaces with Git um application that interfaces with Git um
  application that interfaces with Git um
and does a git status to see what's and does a git status to see what's
  and does a git status to see what's
going on um you could either actually going on um you could either actually
  going on um you could either actually
execute git and set up a git repository execute git and set up a git repository
  execute git and set up a git repository
so that git status outputs what you so that git status outputs what you
  so that git status outputs what you
would like it to Output or you could would like it to Output or you could
  would like it to Output or you could
create you know some way to you know create you know some way to you know
  create you know some way to you know
test configurability to Route around test configurability to Route around
  test configurability to Route around
that and just pretend you output it get that and just pretend you output it get
  that and just pretend you output it get
and give you some mock data and give you some mock data
  and give you some mock data
um we like to actually execute a sub um we like to actually execute a sub
  um we like to actually execute a sub
process but it might not be the binary process but it might not be the binary
  process but it might not be the binary
the real binary that you're executing um the real binary that you're executing um
  the real binary that you're executing um
and so one option is to just execute the and so one option is to just execute the
  and so one option is to just execute the
real thing um in this case get for real real thing um in this case get for real
  real thing um in this case get for real
um and what we do is just ex subprocess um and what we do is just ex subprocess
  um and what we do is just ex subprocess
it there's no real complexity here um it there's no real complexity here um
  it there's no real complexity here um
but we do guard the tests to make sure but we do guard the tests to make sure
  but we do guard the tests to make sure
that the binary exists so um as an that the binary exists so um as an
  that the binary exists so um as an
example like here's something we'll do example like here's something we'll do
  example like here's something we'll do
uh in the in a module in it uh uh in a uh in the in a module in it uh uh in a
  uh in the in a module in it uh uh in a
package in it we'll do a look path to package in it we'll do a look path to
  package in it we'll do a look path to
just kind of as a best guess to see if just kind of as a best guess to see if
  just kind of as a best guess to see if
gits available we'll set some Global gits available we'll set some Global
  gits available we'll set some Global
variable and then in the test themselves variable and then in the test themselves
  variable and then in the test themselves
we'll just we'll just
  we'll just
we'll just guard on it and Skip uh if it we'll just guard on it and Skip uh if it
  we'll just guard on it and Skip uh if it
doesn't exist we do this kind of pattern doesn't exist we do this kind of pattern
  doesn't exist we do this kind of pattern
a lot usually to help with like travisci a lot usually to help with like travisci
  a lot usually to help with like travisci
I testing or testing in environments I testing or testing in environments
  I testing or testing in environments
that can't support certain types of that can't support certain types of
  that can't support certain types of
tests or it's more difficult to support tests or it's more difficult to support
  tests or it's more difficult to support
certain types of tests we do this sort certain types of tests we do this sort
  certain types of tests we do this sort
of of
  of
thing uh the other approach is to mock thing uh the other approach is to mock
  thing uh the other approach is to mock
it and the mocking is where it's a it and the mocking is where it's a
  it and the mocking is where it's a
little bit different because we don't little bit different because we don't
  little bit different because we don't
ever mock the output we actually always ever mock the output we actually always
  ever mock the output we actually always
still execute something but we're still execute something but we're
  still execute something but we're
executing a mock and so to do this the executing a mock and so to do this the
  executing a mock and so to do this the
place where you're subprocessing you place where you're subprocessing you
  place where you're subprocessing you
need to make the exact command arguments need to make the exact command arguments
  need to make the exact command arguments
uh configurable so that you can pass a uh configurable so that you can pass a
  uh configurable so that you can pass a
custom one so that you don't hard code custom one so that you don't hard code
  custom one so that you don't hard code
get in there you don't hard code that get in there you don't hard code that
  get in there you don't hard code that
there's never environmental variables in there's never environmental variables in
  there's never environmental variables in
there just let let the caller somehow there just let let the caller somehow
  there just let let the caller somehow
and again it could be a test only and again it could be a test only
  and again it could be a test only
unexported unexported
  unexported
field but make the caller somehow modify field but make the caller somehow modify
  field but make the caller somehow modify
it um and like I said I found this in it um and like I said I found this in
  it um and like I said I found this in
standard lib it's actually how they test standard lib it's actually how they test
  standard lib it's actually how they test
OS exec um but it's how we test a ton of OS exec um but it's how we test a ton of
  OS exec um but it's how we test a ton of
things so here's here's what it looks things so here's here's what it looks
  things so here's here's what it looks
like it's it's not obvious I think from like it's it's not obvious I think from
  like it's it's not obvious I think from
the beginning but it's also not the beginning but it's also not
  the beginning but it's also not
complicated so what you do is you create complicated so what you do is you create
  complicated so what you do is you create
this helper process thing which returns this helper process thing which returns
  this helper process thing which returns
an exec command which is going to an exec command which is going to
  an exec command which is going to
execute back into the test into a custom execute back into the test into a custom
  execute back into the test into a custom
sort of entry point um what this does is sort of entry point um what this does is
  sort of entry point um what this does is
it runs a really specific test called it runs a really specific test called
  it runs a really specific test called
test helper process it does a dash dash test helper process it does a dash dash
  test helper process it does a dash dash
so we could parse some other stuff later so we could parse some other stuff later
  so we could parse some other stuff later
we also specify an marar to explain that we also specify an marar to explain that
  we also specify an marar to explain that
we really want to run this thing uh and we really want to run this thing uh and
  we really want to run this thing uh and
then we construct a then we construct a
  then we construct a
command and then on the flip side command and then on the flip side
  command and then on the flip side
the actual test itself um what it does the actual test itself um what it does
  the actual test itself um what it does
is it tests if that mbar exists and if is it tests if that mbar exists and if
  is it tests if that mbar exists and if
it doesn't it returns so this this way it doesn't it returns so this this way
  it doesn't it returns so this this way
when you run go test it doesn't actually when you run go test it doesn't actually
  when you run go test it doesn't actually
do anything um otherwise we use that Das do anything um otherwise we use that Das
  do anything um otherwise we use that Das
Dash as a sentinel uh flag to figure out Dash as a sentinel uh flag to figure out
  Dash as a sentinel uh flag to figure out
where the actual args begin so if you where the actual args begin so if you
  where the actual args begin so if you
were doing get status it would actually were doing get status it would actually
  were doing get status it would actually
turn into this go test run something Das turn into this go test run something Das
  turn into this go test run something Das
Dash status is what would actually end Dash status is what would actually end
  Dash status is what would actually end
up happening uh and then we you could do up happening uh and then we you could do
  up happening uh and then we you could do
anything after that you're you're anything after that you're you're
  anything after that you're you're
executing your own program of sorts um executing your own program of sorts um
  executing your own program of sorts um
so we usually switch on like ARG zero so we usually switch on like ARG zero
  so we usually switch on like ARG zero
args left rest and and do stuff and args left rest and and do stuff and
  args left rest and and do stuff and
doing this you could test anything uh doing this you could test anything uh
  doing this you could test anything uh
you could just subprocess test 
  
anything okay so have seven minutes to anything okay so have seven minutes to
  anything okay so have seven minutes to
get through a lot of things so get through a lot of things so
  get through a lot of things so
interfaces um interfaces are important interfaces um interfaces are important
  interfaces um interfaces are important
mocking points I'm sure we've heard this mocking points I'm sure we've heard this
  mocking points I'm sure we've heard this
before they're not just pluggable points before they're not just pluggable points
  before they're not just pluggable points
for consumers they're important mocking for consumers they're important mocking
  for consumers they're important mocking
points for testers um when you have an points for testers um when you have an
  points for testers um when you have an
interface you could interface you could
  interface you could
uh test that interface by creating a uh test that interface by creating a
  uh test that interface by creating a
mock one easily so similar to packages mock one easily so similar to packages
  mock one easily so similar to packages
and functions it's very hard to create and functions it's very hard to create
  and functions it's very hard to create
hard and fast rules of when to create an hard and fast rules of when to create an
  hard and fast rules of when to create an
interface it's very easy to over interface it's very easy to over
  interface it's very easy to over
interface things um you kind of will get interface things um you kind of will get
  interface things um you kind of will get
that over time um but I would say create that over time um but I would say create
  that over time um but I would say create
interfaces where you expect alternate interfaces where you expect alternate
  interfaces where you expect alternate
implementations and create interfaces implementations and create interfaces
  implementations and create interfaces
where um that you feel is the best way where um that you feel is the best way
  where um that you feel is the best way
to test that thing uh but overdoing it to test that thing uh but overdoing it
  to test that thing uh but overdoing it
will complicate the whole will complicate the whole
  will complicate the whole
thing we prefer to use smaller thing we prefer to use smaller
  thing we prefer to use smaller
interfaces where they make sense so even interfaces where they make sense so even
  interfaces where they make sense so even
if we have a big interface if the only if we have a big interface if the only
  if we have a big interface if the only
functionality we we need is the io functionality we we need is the io
  functionality we we need is the io
closer then the function that takes that closer then the function that takes that
  closer then the function that takes that
interface we'd rather just take an IO interface we'd rather just take an IO
  interface we'd rather just take an IO
closer rather than the whole thing um closer rather than the whole thing um
  closer rather than the whole thing um
this might just be a good practice in this might just be a good practice in
  this might just be a good practice in
general but it's a good practice for general but it's a good practice for
  general but it's a good practice for
testing because it simplifies the testing because it simplifies the
  testing because it simplifies the
interface surface era that has to be interface surface era that has to be
  interface surface era that has to be
implemented to do a test so here's an implemented to do a test so here's an
  implemented to do a test so here's an
example um a very common example example um a very common example
  example um a very common example
actually uh common in the standard lib actually uh common in the standard lib
  actually uh common in the standard lib
um is that the serveon type methods take um is that the serveon type methods take
  um is that the serveon type methods take
read write closers even though all if read write closers even though all if
  read write closers even though all if
you look at the things that call serveon you look at the things that call serveon
  you look at the things that call serveon
on their own they're always netc cons on their own they're always netc cons
  on their own they're always netc cons
they're usually always netc cons um but they're usually always netc cons um but
  they're usually always netc cons um but
mocking like I showed you like mocking mocking like I showed you like mocking
  mocking like I showed you like mocking
aeton or even creating like um a small aeton or even creating like um a small
  aeton or even creating like um a small
net con is overkill if you could just net con is overkill if you could just
  net con is overkill if you could just
create a simple read write closer to create a simple read write closer to
  create a simple read write closer to
test the serveon uh function so uh test the serveon uh function so uh
  test the serveon uh function so uh
that's an example of of lowering the that's an example of of lowering the
  that's an example of of lowering the
responsibility of interface that helps responsibility of interface that helps
  responsibility of interface that helps
with testing quite a bit it also makes with testing quite a bit it also makes
  with testing quite a bit it also makes
using it in actual code using it in actual code
  using it in actual code
easier okay next thing is testing as a easier okay next thing is testing as a
  easier okay next thing is testing as a
public API um this is something we only public API um this is something we only
  public API um this is something we only
started doing in the past 18 months or started doing in the past 18 months or
  started doing in the past 18 months or
maybe two years um so newer Hashi cor maybe two years um so newer Hashi cor
  maybe two years um so newer Hashi cor
projects have adopted the practice of projects have adopted the practice of
  projects have adopted the practice of
making a testing or testing undor making a testing or testing undor
  making a testing or testing undor
something file and so you could tell by something file and so you could tell by
  something file and so you could tell by
the name of the file that this becomes the name of the file that this becomes
  the name of the file that this becomes
compiled as part of the actual exported compiled as part of the actual exported
  compiled as part of the actual exported
code it's not just teston code and we code it's not just teston code and we
  code it's not just teston code and we
actually start export we started actually start export we started
  actually start export we started
exporting API exporting API
  exporting API
for the purpose of easing mock creation for the purpose of easing mock creation
  for the purpose of easing mock creation
test writing for consumers of that test writing for consumers of that
  test writing for consumers of that
library and so it allows other people to library and so it allows other people to
  library and so it allows other people to
write tests easier um so here's one write tests easier um so here's one
  write tests easier um so here's one
example which is a config file parser um example which is a config file parser um
  example which is a config file parser um
for terraform for Vault we have some of for terraform for Vault we have some of
  for terraform for Vault we have some of
these which is we export test config these which is we export test config
  these which is we export test config
test config and valid just to easily test config and valid just to easily
  test config and valid just to easily
just give them the structure they need just give them the structure they need
  just give them the structure they need
for valid config and inv valid one um for valid config and inv valid one um
  for valid config and inv valid one um
really basic here's a more complicated really basic here's a more complicated
  really basic here's a more complicated
one that consumers love right it's just one that consumers love right it's just
  one that consumers love right it's just
make me a server um so test server will make me a server um so test server will
  make me a server um so test server will
return the address to connect to as a return the address to connect to as a
  return the address to connect to as a
client and an IO closer to clean up the client and an IO closer to clean up the
  client and an IO closer to clean up the
server uh the place where we have this server uh the place where we have this
  server uh the place where we have this
is Vault Vault actually exports a is Vault Vault actually exports a
  is Vault Vault actually exports a
function for you Ingo to create a fully function for you Ingo to create a fully
  function for you Ingo to create a fully
inmemory non-durable server that's that inmemory non-durable server that's that
  inmemory non-durable server that's that
is Vault so you could actually create a is Vault so you could actually create a
  is Vault so you could actually create a
vault client connect to it and test vault client connect to it and test
  vault client connect to it and test
communicating with it and it's a communicating with it and it's a
  communicating with it and it's a
publicly supported exported publicly supported exported
  publicly supported exported
API um and then the other way is we API um and then the other way is we
  API um and then the other way is we
actually export functions that test actually export functions that test
  actually export functions that test
implementations of an interface to see implementations of an interface to see
  implementations of an interface to see
if they behave to spec of what our if they behave to spec of what our
  if they behave to spec of what our
Behavior expects so example is we have a Behavior expects so example is we have a
  Behavior expects so example is we have a
library that downloads things there's no library that downloads things there's no
  library that downloads things there's no
other way to describe it than it just other way to describe it than it just
  other way to describe it than it just
downloads things from anywhere using downloads things from anywhere using
  downloads things from anywhere using
anything um and so you implement this anything um and so you implement this
  anything um and so you implement this
downloader interface in order to uh downloader interface in order to uh
  downloader interface in order to uh
create a new downloader and there's some create a new downloader and there's some
  create a new downloader and there's some
Behavior we expect like if we download Behavior we expect like if we download
  Behavior we expect like if we download
um if we download without the um if we download without the
  um if we download without the
destination directory existing we expect destination directory existing we expect
  destination directory existing we expect
you to create that directory we can't you to create that directory we can't
  you to create that directory we can't
represent that in Go's type system so represent that in Go's type system so
  represent that in Go's type system so
that's an implementation detail that's that's an implementation detail that's
  that's an implementation detail that's
easy to miss when you're implementing a easy to miss when you're implementing a
  easy to miss when you're implementing a
thing implenting a downloader and so we thing implenting a downloader and so we
  thing implenting a downloader and so we
actually export a function called test actually export a function called test
  actually export a function called test
downloader that runs through a bunch of downloader that runs through a bunch of
  downloader that runs through a bunch of
table tests for any generic downloader table tests for any generic downloader
  table tests for any generic downloader
to ensure that the behavior is what we to ensure that the behavior is what we
  to ensure that the behavior is what we
expect um we also export things like expect um we also export things like
  expect um we also export things like
mock structures so that if you wanted to mock structures so that if you wanted to
  mock structures so that if you wanted to
pass a downloader Mock and test that pass a downloader Mock and test that
  pass a downloader Mock and test that
it's downloading the right thing you it's downloading the right thing you
  it's downloading the right thing you
have have
  have
that uh the last thing is uh we I have that uh the last thing is uh we I have
  that uh the last thing is uh we I have
this this package called go testing this this package called go testing
  this this package called go testing
interface and what it does is it creates interface and what it does is it creates
  interface and what it does is it creates
a testing. t it actually is called that a testing. t it actually is called that
  a testing. t it actually is called that
interface um and we you we use that for interface um and we you we use that for
  interface um and we you we use that for
the test helpers because if you actually the test helpers because if you actually
  the test helpers because if you actually
import the real testing package that import the real testing package that
  import the real testing package that
adds Flags to the global flag thing and adds Flags to the global flag thing and
  adds Flags to the global flag thing and
so if if if the the package importing so if if if the the package importing
  so if if if the the package importing
your own package uses the global Flags your own package uses the global Flags
  your own package uses the global Flags
then they'll all a sudden I'll have then they'll all a sudden I'll have
  then they'll all a sudden I'll have
these run and so on flags introduced and these run and so on flags introduced and
  these run and so on flags introduced and
that's annoying so if you use the that's annoying so if you use the
  that's annoying so if you use the
interface uh then you can pass it on in interface uh then you can pass it on in
  interface uh then you can pass it on in
so here's an example of that test config so here's an example of that test config
  so here's an example of that test config
we import this we use the interface you we import this we use the interface you
  we import this we use the interface you
can see it's missing the asterisk um you can see it's missing the asterisk um you
  can see it's missing the asterisk um you
could still use it like it's a testing. could still use it like it's a testing.
  could still use it like it's a testing.
t and you can pass a real testing. into t and you can pass a real testing. into
  t and you can pass a real testing. into
it okay we're getting there um custom it okay we're getting there um custom
  it okay we're getting there um custom
Frameworks so uh go test is a really Frameworks so uh go test is a really
  Frameworks so uh go test is a really
incredible workflow tool so when we do incredible workflow tool so when we do
  incredible workflow tool so when we do
things that aren't quite unit test we things that aren't quite unit test we
  things that aren't quite unit test we
still try to fit it still try to fit it
  still try to fit it
into I don't know are they kicking I into I don't know are they kicking I
  into I don't know are they kicking I
still have time still have time
  still have time
um I'll keep talking though so we still um I'll keep talking though so we still
  um I'll keep talking though so we still
try to fit these tests that don't uh oh try to fit these tests that don't uh oh
  try to fit these tests that don't uh oh
maybe it's maybe it's my fault actually maybe it's maybe it's my fault actually
  maybe it's maybe it's my fault actually
my computer may have 
  
died I don't know um I'll just talk died I don't know um I'll just talk
  died I don't know um I'll just talk
through that one and then I'll probably through that one and then I'll probably
  through that one and then I'll probably
have to end after that one because I have to end after that one because I
  have to end after that one because I
don't have the slides anymore um but don't have the slides anymore um but
  don't have the slides anymore um but
when we build things that don't quite when we build things that don't quite
  when we build things that don't quite
fit unit tests we still try to build um fit unit tests we still try to build um
  fit unit tests we still try to build um
build them into the go test workflow so build them into the go test workflow so
  build them into the go test workflow so
a good example with this is ter form has a good example with this is ter form has
  a good example with this is ter form has
an acceptance test Library what the an acceptance test Library what the
  an acceptance test Library what the
acceptance test Library does is it takes acceptance test Library does is it takes
  acceptance test Library does is it takes
real terraform configurations and then real terraform configurations and then
  real terraform configurations and then
from the real terraform configurations from the real terraform configurations
  from the real terraform configurations
it creates real infrastructure and it's it creates real infrastructure and it's
  it creates real infrastructure and it's
black boxy in a way it's mostly blackbox black boxy in a way it's mostly blackbox
  black boxy in a way it's mostly blackbox
um and we we run these every night um we um and we we run these every night um we
  um and we we run these every night um we
spin up thousands of resources on spin up thousands of resources on
  spin up thousands of resources on
something like 50 different providers something like 50 different providers
  something like 50 different providers
every night um and we use go tests to every night um and we use go tests to
  every night um and we use go tests to
trigger this even though they're not trigger this even though they're not
  trigger this even though they're not
unit tests and so what we did was we unit tests and so what we did was we
  unit tests and so what we did was we
built our own framework that we call built our own framework that we call
  built our own framework that we call
from a test um that has its own API from a test um that has its own API
  from a test um that has its own API
doesn't look like a unit test at all doesn't look like a unit test at all
  doesn't look like a unit test at all
doesn't call fatal F or any of those doesn't call fatal F or any of those
  doesn't call fatal F or any of those
things um it just basically is a things um it just basically is a
  things um it just basically is a
structure that says here are the config structure that says here are the config
  structure that says here are the config
files to run here are the credentials files to run here are the credentials
  files to run here are the credentials
for the cloud platform here are some you for the cloud platform here are some you
  for the cloud platform here are some you
know tests to run like you should be know tests to run like you should be
  know tests to run like you should be
able to reach this IP um this if you run able to reach this IP um this if you run
  able to reach this IP um this if you run
this API call then it should return an this API call then it should return an
  this API call then it should return an
instance we we do that and then we run instance we we do that and then we run
  instance we we do that and then we run
it with a special Flag by introducing it with a special Flag by introducing
  it with a special Flag by introducing
Flags like I showed you earlier like the Flags like I showed you earlier like the
  Flags like I showed you earlier like the
update flag we actually have like an update flag we actually have like an
  update flag we actually have like an
acceptance flag and so when you run go acceptance flag and so when you run go
  acceptance flag and so when you run go
test with acceptance flag then it runs test with acceptance flag then it runs
  test with acceptance flag then it runs
the acceptance sest Suite skips all the the acceptance sest Suite skips all the
  the acceptance sest Suite skips all the
rest uh and we get features like that in rest uh and we get features like that in
  rest uh and we get features like that in
in Vault it's the same way we in Vault in Vault it's the same way we in Vault
  in Vault it's the same way we in Vault
we have an acceptance test framework for we have an acceptance test framework for
  we have an acceptance test framework for
um testing secret backends off backends um testing secret backends off backends
  um testing secret backends off backends
and all these different things same and all these different things same
  and all these different things same
thing we built a custom framework to thing we built a custom framework to
  thing we built a custom framework to
make writing those easy um but you still make writing those easy um but you still
  make writing those easy um but you still
execute them with go test it just as a execute them with go test it just as a
  execute them with go test it just as a
workflow is what we try to do um I think workflow is what we try to do um I think
  workflow is what we try to do um I think
there's only two more sections I'm going there's only two more sections I'm going
  there's only two more sections I'm going
to put the slides online so um I'll end to put the slides online so um I'll end
  to put the slides online so um I'll end
there and I'm out of time anyway so there and I'm out of time anyway so
  there and I'm out of time anyway so
thank you thank you
  thank you
[Applause]
