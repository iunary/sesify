# SESIFY

Simple email sender cli currently support amazon ses, customizable templates and attachments.

# Usage

```bash
sesify send --from john@example.com --recipients list.csv --template templates/template.html --subject "hello there"
```

## Template
Template support the following recipient varibales:

```golang
{{.UUID}}
{{.Firstname}}
{{.Lastname}}
{{.Email}}
```


# TODO

- [-] Add attachments
- [-] Tests