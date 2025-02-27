# Roadmap

- [ ] In addition to kubernetes events, expose the ability to select on a status
- [ ] Create helm chart
- [ ] Expose metrics and traces via OTEL
- [ ] Have EventTriggeredJob create a meta block (inherit xyz from parent, like annotations, namespace, etc.)
- [ ] Create a Prometheus EventProducer

---

- [ ] Decouple event consumption from job dispatching
- [ ] Decouple eventSelector and jobTemplate into separate CRDs (eventSelector should actually be of type EventProducer called kubernetes-event. There can be multiple resourceEventProducers with options (kind, namePattern, namespacePattern, labelSelector, eventTypes), and should communicate with kubeevent via a webhook. Other EventProducers could be a Prometheus EventProducer, Github EventProducer, etc., but in this project we would just have kubernetes-event, kubernetes-resource-status, and prometheus)
