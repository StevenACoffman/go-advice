## Alerts

Adapted from [Rob Ewaschuk's chapter in Google's Site Reliability Engineering](https://docs.google.com/document/d/199PqyG3UsyXlwieHaqbGiWVa8eMWi8zzAn0YfcApr8Q/edit#)

The underlying point is to create a system that still has accountability for responsiveness, but doesn't have the high cost of waking someone up.

### Summary
When you are auditing or writing alerting rules, consider these things to keep your oncall rotation happier:

+ Alerts should be urgent, important, actionable, and real.
+ They should represent either ongoing or imminent problems with your service.
+ Err on the side of removing noisy alerts – over-monitoring is a harder problem to solve than under-monitoring.
+ You should almost always be able to classify the problem into one of: availability & basic functionality; latency; correctness (completeness, freshness and durability of data); and feature-specific problems.
+ Symptoms are a better way to capture more problems more comprehensively and robustly with less effort.
+ Include cause-based information in symptom-based pages or on dashboards, but avoid alerting directly on causes.
+ The further up your serving stack you go, the more distinct problems you catch in a single rule.  But don't go so far you can't sufficiently distinguish what's going on.
+ If you want a quiet oncall rotation, it's imperative to have a system for dealing with things that need timely response, but are not imminently critical.

### Playbooks

Playbooks (or runbooks) are an important part of an alerting system; it's best to have an entry for each alert or family of alerts that catch a symptom, which can further explain what the alert means and how it might be addressed. The best playbooks I've seen have a few notes about exactly what the alert means, and what's currently interesting about an alert ("We've had a spate of power outages from our widgets from VendorX; if you find this, please add it to Bug 12345 where we're tracking things for patterns".)  Most such notes should be ephemeral, so a wiki or similar is a great tool.

Matthew Skelton & Rob Thatcher have an excellent [run book template](https://github.com/SkeltonThatcher/run-book-template). This template will help teams to fully consider most aspects of reliably operating most interesting software systems, if only to confirm that "this section definitely does not apply here" - a valuable realization.

### Tracking & Accountability
Track your high priority alerts (especially pager duty pages if you have them), and all your other alerts.  If an alert is firing and people just say "I looked, nothing was wrong", that's a pretty strong sign that you need to remove the alert rule, or demote it or collect data in some other way.  Alerts that are less than 50% accurate are broken; even those that are false positives 10% of the time merit more consideration.

Having a system in place (e.g. a weekly review of all alerts, and quarterly statistics) can help keep a handle on the big picture of what's going on, and tease out patterns that are lost when the alert is handed from one human to the next.


