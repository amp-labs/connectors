# Description

Visit `https://docs.lob.com/` and open the browser's developer console.

Run the following command to copy the OpenAPI specification to the clipboard:

```js
copy(JSON.stringify(__redoc_state.spec.data, null, 2))
```

The specification is embedded in the page as the `__redoc_state` variable.

If the variable is not available in the console, locate the `<script>` element containing it and evaluate that script first. For example, at the moment, it is the script at index `22`:

```js
eval(document.querySelectorAll("script")[22].innerText)
```

Then copy the OpenAPI specification:

```js
copy(JSON.stringify(__redoc_state.spec.data, null, 2))
```

> **Note:** The script index may change when the documentation site is updated. If `__redoc_state` is already available in the console, there is no need to locate and evaluate the script manually.
