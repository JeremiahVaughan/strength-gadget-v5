Style scoping for base.html and <page-name>.html is done implicitly since only one page template and one base template will only ever be active at one time.

Component level style scoping is to be performed via naming convention. For example, the button.html component must have each of its css classes prefixed with the name of the component file: "button-content", "button-color", etc.

This naming convention also applies to template names, for example button-content and button-styling.

This naming convention also applies to element ids.

