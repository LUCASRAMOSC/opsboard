# OpsBoard — Product Scope

## 1. Product vision

OpsBoard is a journey-first operational dashboard that translates technical service health into user and business impact.

The project is designed around three main questions:

1. What is failing?
2. What user journeys are affected?
3. What changed before the degradation started?

OpsBoard is not intended to replace traditional infrastructure monitoring or APM platforms.

Its purpose is to provide a simpler operational context layer for development teams.

## 2. Target scenario

The initial version of OpsBoard will use a fictional e-commerce platform called **ShopFlow** as its demonstration environment.

ShopFlow represents a small production system composed of multiple technical services:

- Web Frontend
- Auth API
- Orders API
- Payments API
- PostgreSQL

These services support user-facing business journeys such as:

- Customer Login
- Checkout
- Order Tracking

A business journey may depend on multiple services.

For example, the **Checkout** journey depends on:

- Web Frontend
- Orders API
- Payments API
- PostgreSQL

If the Payments API becomes degraded or unavailable, OpsBoard should identify the affected technical service and propagate that degradation to the Checkout journey.

The dashboard should then provide operational context by showing:

- The affected service
- The impacted business journey
- The severity of the impact
- Recent changes or deployments that may be related to the degradation

ShopFlow exists only as demonstration data and does not represent a real company or production environment.

## 3. Core domain concepts

### Workspace

A Workspace represents an OpsBoard environment and groups all operational data related to a product or system.

The initial public demo will use a fictional workspace called **ShopFlow**.

A Workspace contains:

- Services
- Business Journeys
- Health Events
- Deployments
- Incidents

The Workspace concept allows OpsBoard to support multiple isolated environments in the future without coupling the domain model to a single application.

### Service

A Service represents a technical component that participates in one or more business journeys.

Examples include:

- Frontend applications
- APIs
- Databases

Each service has a criticality level that represents its operational importance.

Initial criticality levels are:

- Low
- Medium
- High
- Critical

The current operational status of a service should be derived from its most recent Health Event instead of being treated as an independent source of truth.

Initial service statuses are:

- Healthy
- Degraded
- Down

### Business Journey

A Business Journey represents an action or flow that is meaningful to the user or business.

Examples include:

- Customer Login
- Checkout
- Order Tracking

A Business Journey depends on one or more Services, and the same Service may participate in multiple Business Journeys.

Each Business Journey also has a criticality level that contributes to the calculation of operational impact.

### Health Event

A Health Event represents an observation of a Service at a specific point in time.

The initial version will record:

- Service status
- Response time
- Observation timestamp

Health Events provide the historical information required to understand when a degradation started and how service health changed over time.

### Deployment

A Deployment represents a change released to a Service.

The initial version will include:

- Version
- Commit identifier
- Environment
- Deployment timestamp

Deployments provide operational context that allows OpsBoard to identify recent changes that may be correlated with a service degradation.

A correlation must never be presented as confirmed causation.

### Incident

An Incident represents a recognized operational problem.

An Incident is different from a Health Event.

A Health Event describes an observed technical state, while an Incident represents an operational problem being tracked and investigated.

An Incident may affect one or more Services, and a Service may participate in multiple Incidents over time.

Initial incident states are:

- Investigating
- Identified
- Monitoring
- Resolved

### Impact Analysis

Impact Analysis is the core domain capability responsible for translating technical service degradation into user and business impact.

When a Service becomes degraded or unavailable, OpsBoard should:

1. Identify the affected Service.
2. Find the Business Journeys that depend on that Service.
3. Consider the criticality of the affected Service and Business Journeys.
4. Determine the severity of the operational impact.
5. Look for recent Deployments related to the affected Service.
6. Present possible deployment correlation when relevant.

The first version should use a deterministic and explainable impact calculation rather than an opaque or AI-based model.

### Core relationships

The initial domain contains two important many-to-many relationships:

- A Business Journey may depend on multiple Services, and a Service may support multiple Business Journeys.
- An Incident may affect multiple Services, and a Service may be associated with multiple Incidents over time.

## 4. MVP scope

The first version of OpsBoard should provide a complete demonstration of the journey-first operational impact concept without attempting to become a full observability platform.

### Service management

The user should be able to:

- View the Services available in a Workspace.
- View the current operational status of each Service.
- View the criticality of each Service.
- View recent Health Events for a Service.

The initial operational statuses are:

- Healthy
- Degraded
- Down

### Business Journeys

The user should be able to:

- View the Business Journeys available in a Workspace.
- View the criticality of each Business Journey.
- See which Services each Business Journey depends on.
- Understand the current operational condition of a Business Journey based on its dependencies.

### Health Events

OpsBoard should maintain a basic history of Service health observations.

Each Health Event should contain:

- Service
- Operational status
- Response time
- Observation timestamp

Health Events will be used to determine the current Service status and provide basic health history.

### Deployments

The user should be able to view recent Deployments associated with Services.

Each Deployment should initially contain:

- Service
- Version
- Commit identifier
- Environment
- Deployment timestamp

Deployment data will be used by OpsBoard to identify possible correlations between recent changes and Service degradation.

### Incidents

The user should be able to:

- View active Incidents.
- View resolved Incidents.
- See which Services are affected by an Incident.
- See the Incident severity and current state.

The initial Incident states are:

- Investigating
- Identified
- Monitoring
- Resolved

### Impact Analysis

When a Service becomes degraded or unavailable, OpsBoard should automatically determine:

1. Which Business Journeys depend on the affected Service.
2. Which Business Journeys are currently impacted.
3. The severity of the operational impact.
4. A deterministic and explainable impact score.
5. Whether a recent Deployment may be correlated with the degradation.

The Impact Analysis should clearly explain why a Business Journey is considered affected.

### Incident simulation

The public demo should allow visitors to simulate predefined operational scenarios.

Initial scenarios may include:

- Payments API high latency.
- Payments API unavailable.
- Auth API degradation.

A simulation should create Health Events and allow the visitor to immediately observe how the degradation affects Business Journeys and the calculated impact.

### Demo reset

The public demo should provide a way to restore the demonstration environment to its initial state.

This allows visitors to experiment with the application without permanently changing the sample data.

## 5. MVP non-goals

The first version of OpsBoard is intentionally focused on operational impact rather than full infrastructure observability.

The following capabilities are explicitly outside the MVP scope:

### Infrastructure monitoring

OpsBoard will not initially monitor:

- CPU usage
- Memory usage
- Disk usage
- Network traffic
- Operating system metrics
- Physical or virtual hosts

### Advanced observability

The MVP will not include:

- Log aggregation
- Distributed tracing
- Application Performance Monitoring (APM)
- OpenTelemetry integration
- Prometheus integration
- Grafana integration
- Custom metric ingestion

### Infrastructure orchestration

The MVP will not include:

- Kubernetes monitoring
- Container orchestration
- Infrastructure provisioning
- Infrastructure as Code
- Cloud resource management

### Advanced event processing

The MVP will not require:

- Message brokers
- Event streaming platforms
- Redis
- Kafka
- RabbitMQ

The initial architecture should remain simple and appropriate for the expected workload.

### Artificial intelligence

The MVP will not use AI or machine learning to determine incident causes or impact.

Impact Analysis and deployment correlation should use deterministic and explainable rules.

### Arbitrary external monitoring

The public demo will not allow visitors to provide arbitrary URLs or network targets for monitoring.

Operational scenarios in the public demo will use controlled and predefined data to avoid security and abuse risks.

### Authentication and authorization

Full user authentication, roles and permissions are not required for the first demonstration of the core product concept.

Authentication may be introduced in a later version when it provides meaningful value to the product.

### Real-time requirements

The MVP does not require WebSockets or real-time streaming infrastructure.

The application may initially use standard HTTP requests and periodic data refresh when necessary.
