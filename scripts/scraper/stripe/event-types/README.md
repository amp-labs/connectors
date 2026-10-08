# Query

Visit https://docs.stripe.com/api/events/types.

Run in console:
```
[...document.querySelectorAll("#api-section-event_types > div.ApiSectionGrid.algolia-is-current-method.⚙.as-61.as-62.as-63.as-64.as-65.as-66.as-67.as-68.as-69.⚙182fbad > div > div > ul h4")].map(el => el.innerText)
```

Manually modify output to produce `events.json.

# Script

Run `main.go` to get report on events.
It produces the `mapping.json` which can be used in the connector to map Ampersand Object subscriptions of
`create`, `update`, `delete` to Stripe event names.