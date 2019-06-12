Importing a Helm Chart
================

To import an existing checked-out Helm chart into Replicated, you can use 

    replicated release import-helm-chart ${CHART_ROOT}
    
where `CHART_ROOT` is the path on your filesystem to a Helm chart archive or repository
containing a `Chart.yaml` and default `values.yaml`.


This will output a recommended release to STDOUT. To store this in `replicated.yaml`, pipe the output to


### Maturity

As of version 0.14.0 (August 2019), this is a fairly rough tool. It is not regularly tested on a large corpus of helm charts.
The primary goal of this tool is to get you 80-90% of the way to bundling your helm-based application in Replicated, and while
the results may be immediately usable out of the box, it's likely you'll need to make some small tweaks to make the outputs of 
`import-helm-chart` production-ready. Some examples of changes you might want to make:

- Mark some config options as `hidden`, or remove them entirely and just use defaults in your `values.yaml`  (see [advanced usage](#advanced-usage-prepare-helm-values))
- Add descriptions or other defaults to your `config` items
- Reworking some of your helm `templates` so they play nicer with a string-based templating of boolean and numeric values.

It should also be noted that certain values techniques are currently not supported at all, including lists of objects (e.g. [tolerations](https://github.com/helm/charts/blob/master/stable/prometheus/values.yaml#L130) and other subjects of the sprig `toYaml` construct provided by helm)

How it works
--------------

The import tool is essentially an end-to-end implementation of [How can I ship a Helm application as a Replicated appliance?](https://help.replicated.com/community/t/how-can-i-ship-a-helm-application-as-a-replicated-appliance/162). It requires `helm` to be installed on the machine where it is run. It essentially runs the equivalent of:

```sh
helm init --client-only
helm dependency update
cat values-replicated.yaml | helm template -f - ${CHART_ROOT}
```

### Translating values

The `values-replicated.yaml` is generated during this process, and will contain values that will translate `{{ .Values.foo }}` in a chart into `{{repl ConfigOption "foo"}}` in the final artifact. See [advanced usage](#advanced-usage-prepare-helm-values) to build and customize this values file yourself.

For example, given the values file

```yaml
foo: bar
postgres:
  replicas: 2
```

and (redacted-for-brevity) k8s YAML

```yaml
kind: Deployment
spec:
  replicas: {{.Values.postgres.replicas}}
  template:
    # ...
    env:
     - name: FOO
       value: {{.Values.foo}} 
```

this tool will generate a values file

```yaml

foo: '{{repl ConfigOption "foo"}}'
postgres:
  replicas: '{{repl ConfigOption "postgres.replicas"}}'
```

and the final Deployment sent to replicated will look something like
```yaml

kind: Deployment
spec:
  replicas: {{repl ConfigOption "postgres.replicas"}}
  template:
    # ...
    env:
     - name: FOO
       value: {{repl ConfigOption "foo"}}
```

and the Replicated Config Section that describes the on-prem user config screen will look like


```yaml
config:
- name: helm_values
  items:
    - name: foo
      title: foo
      type: text
      default: bar
    - name: postgres.replicas
      title: postgres.replicas
      type: text
      default: 2
```

### Advanced usage: `prepare-helm-values`

As mentioned in [translating values](#translating-values), a replicated-specific values file will be generated
as part of the process. If you'd like to reduce the number of values exposed to the end customer, or customize messaging and defaults, you can regenerate both the `values-replicated.yaml` and the Replicated config screen YAML with

    replicated release prepare-helm-values ${PATH_TO_VALUES_YAML}
    
This will print (to stdout) the intermediate Values representation, and the config screen starter for you to customize.
