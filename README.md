# Kubevent

Kubevent is a Kubernetes controller that allows you to trigger a [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) based on a Kubernetes event (eventually, any generic event sent to the Kubevent webhook).

This project is inspired by policy engines like [Kyverno](https://kyverno.io/) and operators like [Metacontroller](https://metacontroller.github.io/metacontroller/intro.html) which allow for flexible controller-like logic as versioned Kubernetes resources.

This tool, however, is solely used to trigger a Job when an event happens in an opinionated fashion. Kyverno and Metacontroller can be used for similar purposes but are significantly heavier dependencies.

## Using Kubevent: the EventTriggeredJob

Kubevent introduces the `EventTriggeredJob` CRD to help associate Jobs with events:

```yaml
apiVersion: kubevent.roshanbhatia.com/v1alpha1
kind: EventTriggeredJob
metadata:
  name: example-job-template
spec:
  eventSelector:
    resourceKind: "Pod"
    namePattern: "web-*"
    namespacePattern: "prod-*"
    labelSelector:
      matchLabels:
        app: myapp
    eventTypes: ["CREATE", "DELETE"]
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            command: ["echo", "Hello from Kubevent!"]
          restartPolicy: Never
```

This CRD allows you to define:

- Which resources to watch (by kind, name pattern, namespace pattern, labels)
- Which event types should trigger a job (CREATE, UPDATE, DELETE)
- The job template to execute when an event is triggered

## Local Development

### Requirements

- [Go 1.21+](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Kind](https://kind.sigs.k8s.io/)
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)
- [chainsaw](https://kyverno.github.io/chainsaw/0.2.3/)

### Dev Tooling

Run `make help` to list all available make targets for local development and testing.

## Contributing

I don't actively watch this repo but feel free to fork and do what you desire with it.
