# Go Developer Advice

This repo contains a curated distillation of the best advice I have encountered, and that I give to any aspiring (Go) developers I mentor.

The [./summary_rules.md](./summary_rules.md) is distilled from:
+ [Ben Johnson](./Sources/benbjohnson_rules.md), from his defunct gobeyond.dev site: 
  + [crud](./Sources/benbjohnson/crud.md)
  + [failure-is-your-domain.md](./Sources/benbjohnson/failure-is-your-domain.md)
  + [packages-as-layers.md](./Sources/benbjohnson/packages-as-layers.md)
  + [real-world-sql-part-one.md](./Sources/benbjohnson/real-world-sql-part-one.md)
  + [standard-package-layout.md](./Sources/benbjohnson/standard-package-layout.md)
  + [structuring-applications-in-go.md](./Sources/benbjohnson/structuring-applications-in-go.md)
  + [structuring-tests-in-go.md](./Sources/benbjohnson/structuring-tests-in-go.md)
  + [wtf-dial.md](./Sources/benbjohnson/wtf-dial.md)
+ [Gary Bernhardt](./Sources/gbernhardt_rules.md) from:
  + [functional core / imperative shell](./Sources/gbernhardt/bernhardt-functional-core-imperative-shell.md)
+ [Mat Ryer](./Sources/matryer_rules.md) from:
  + [How I Write HTTP Services in Go after 13 years](https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years/)
+ [Michael Feathers](./Sources/mfeathers_rules.md) from:
  + [YOW 2016 Testing Patience](./Sources/mfeathers/yow-2016-michael-feathers-testing-patience.md)
  + [Geekfest Testing Patience](./Sources/mfeathers/geekfest-michael-feathers-testing-patience.md)
  + [Testing Patience Slides](./Sources/mfeathers/Feathers-TestingPatience.pdf)
+ [Mitchell Hashimoto](./Sources/mitchellh_rules.md) from:
  + [GopherCon 2017 Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk) 
  + [GopherCon 2017 Advanced Testing with Go Slides](./Sources/mitchellh/Advanced_Testing_with_Go.pdf)
+ [Martin Kleppmann](./Sources/kleppmann_rules.md) from:
  + [Designing Data-Intensive Applications, 2nd Edition](https://www.oreilly.com/library/view/designing-data-intensive-applications/9781098119058/)
+ [John Ousterhout](./Sources/ousterhout_rules.md) from:
  + [A Philosophy of Software Design, 2nd Edition](https://www.amazon.com/dp/173210221X) 

The following resources are highly recommended, but are not yet incorporated into the distillation.

### Data Science Resources

There's a HUGE component that is more psychological: e.g., how to ask "good" questions, how to design data products with the end user in mind. One perspective on DS that has always resonated with me is that it's fundamentally about getting really good at the following translational cycle of converting:

> business question -> data question -> data answer -> business answer

Doing that cycle well is hard. For instance, business questions are often quite abstract (like "is X working?"). So before we can even begin to write a SQL query to curate numbers for crunching, there's a whole host of work to operationalize the question. For instance:

What does it mean for X to be working... or not? What is the theory of action? Are the user's implicitly/explicitly incentivized to use it as intended? What are the various features? Is the data sufficiently rich that I can recover the full set of possible user behaviors? Are there readily available measures that provide meaningful insight, or do we need to develop/validate measures that isolate signal from noise. For instance, sometimes simple behavioral measures like pageviews or time on task are meaningful... but sometimes we need bespoke measures, like trying to understand the depth/substance of student conversations with X. Do the decision makers have preexisting beliefs/biases that will influence how they receive analytic insights? How can we develop the measurement approach and present the insights in a way that is most palatable, given these beliefs?

That type of thinking is a bit harder to develop than, say, learning SQL basics. It requires business/domain-specific knowledge, as well as some general psychology/human-centered design type thinking. But there are certainly resources that focus on this. I share some below. In general, it's "easy" enough to learn a new technical tool with a bit of time and focused effort (oh this company uses a different SQL dialect than I'm used to. I need to learn the nuances.) But learning to think like a researcher/DS, to ask good questions, to get to the "why" of it all, and to present insights in an impactful and digestible way involves so much more than just writing a query. (And to be fair, there are plenty of folks with a DS title who are still developing these skills, including me).

Once we've begun to operationalize a business question and worked through some of the psychological stuff above, then comes the data modeling, exploratory data analysis (EDA), statistical analysis, dashboard design, etc. These are more technical, but again still intimately tied to the psychological. For instance, in my experience, data modeling for eng/infra is different from data modeling for data science/research. In the latter case, I often want to over-curate data to support rich EDA. For instance, if I see something odd in a top-level analysis, I want to be able to quickly drill down to see if it's a real pattern or an error. Perhaps I want to be able to segment by various features or condition on some metadata. Rich EDA is SOO useful, so I am always thinking ahead when developing the data model:

What types of (un)expected patterns might we observe? if I see the expected pattern, is that because the product is working... or because there's some bias in the data (like stronger students being more likely to use the product)? what data can I curate in advance to distinguish true patterns from confounding. What are the various segments of users? Do I expect similar patterns across each segment? How do I balance thoroughness (e.g., curating enough data to support rich analysis to rule out alternate explanations) vs. efficiency (e.g., getting quick directional insights)?

I like to ask “what action will you take with this information?” Been asking that for 25+ years.
In general, I find it’s best task open-ended questions that cause a person to explain what they’re thinking, and in the process justify the answer to themselves.


##### "The Art of Data Science" (Roger D. Peng & Elizabeth Matsui, free online book)

Emphasizes exploratory analysis as a process of discovery rather than a mechanical task.
Other stuff by Roger Peng is also really good: like the Not So Standard Deviations podcast and Exploratory Data Analysis with R


##### "Storytelling with Data" (2015) – Cole Nussbaumer Knaflic

Helps data professionals communicate insights effectively, considering audience psychology and cognitive load.


##### "How to Measure Anything" (2007) – Douglas Hubbard

Teaches how to frame business problems as measurable data questions, even when dealing with uncertainty.


##### "The Data Detective" (2021) – Tim Harford

Focuses on the cognitive biases and psychology of interpreting data, helping data professionals think critically about how they (and others) make sense of numbers.


##### "How Charts Lie" (2019) – Alberto Cairo

Explores how people misinterpret visualizations, helping data scientists design for clarity and minimize bias.


##### "Mostly Harmless Econometrics" (Angrist & Pischke, 2008) - [free pdf]

This is a statistics textbook. But it's fairly accessible. And it's really good for thinking critically about causality vs. correlation (especially in business contexts where we aren't always running causal experiments).

##### “Never Split the Difference” – Chris Voss

A FBI Hostage negotiator describes how to ask better questions and figure out what people really need


