# OpsBoard — Initial Architecture

## 1. Architecture overview

OpsBoard starts with a modular monolithic backend, a separate frontend application and a relational database.

The initial architecture is intentionally simple and designed to support the MVP without introducing infrastructure that is not yet necessary.

```text
React + TypeScript
        ↓
      REST API
        ↓
      Go + Gin
        ↓
    PostgreSQL
```

This architecture represents the initial state of the project and may evolve as new requirements appear.

New architectural components should only be introduced when they solve a concrete product or technical problem.

## 2. Monorepo structure

OpsBoard uses a single repository for application code, infrastructure configuration and project documentation.

The initial structure is:

```text
opsboard/
├── apps/
│   ├── api/
│   └── web/
├── docs/
├── infra/
│   └── nginx/
├── .editorconfig
├── .env.example
├── .gitignore
├── docker-compose.yml
├── LICENSE
└── README.md
```

The monorepo keeps the frontend, backend and infrastructure changes visible within the same project history.

## 3. Frontend responsibilities

The frontend will be implemented with React and TypeScript.

Its main responsibilities are:

- Present Business Journey health and operational impact
- Present Service health and recent Health Events
- Display Incidents and Deployments
- Allow visitors to trigger predefined demo simulations
- Explain Impact Analysis results
- Communicate with the backend through HTTP requests

Business Journeys should remain more prominent than low-level technical information.

The frontend should not contain the authoritative Impact Analysis rules.

## 4. Backend responsibilities

The backend will be implemented with Go and Gin.

It is responsible for:

- Exposing the REST API
- Managing the OpsBoard domain
- Applying business rules
- Persisting and retrieving data
- Creating Health Events during demo simulations
- Calculating operational impact
- Identifying possible deployment correlations
- Enforcing demo isolation and interaction boundaries

Business rules should remain independent from HTTP-specific concerns whenever practical.

The backend will initially run as a single application instead of multiple services.

## 5. Database responsibilities

PostgreSQL will be the primary persistence layer.

It will store data related to:

- Workspaces
- Services
- Business Journeys
- Journey and Service relationships
- Health Events
- Deployments
- Incidents
- Incident and Service relationships

The Service operational status should be derived from its most recent Health Event instead of being maintained as a separate authoritative value.

Database design should follow the domain model without introducing abstractions that are not required by the MVP.

## 6. Local development architecture

The project should be runnable locally through Docker Compose.

The initial local environment is expected to contain:

```text
Browser
   ↓
Nginx
   ├── Web
   └── API
        ↓
    PostgreSQL
```

Docker Compose should provide a reproducible development environment for the application services.

Nginx will act as the entry point for the containerized local environment and route requests to the appropriate application component.

Individual applications may still be executed directly during development when that provides a faster development workflow.

## 7. Public demo architecture

The public demo may use a different deployment topology from the local environment while keeping the same application boundaries.

The initial target is:

```text
Visitor
   ↓
Cloudflare Pages
React + TypeScript
   ↓
HTTPS
   ↓
Containerized Go API
   ↓
Managed PostgreSQL
```

The frontend may be deployed independently as a static application.

The Go API should remain containerized so the same application can be executed consistently across environments.

The exact hosting providers may change if operational or free-tier constraints change.

## 8. Architectural boundaries

The initial architecture follows these boundaries:

- Frontend presentation logic should not become the source of truth for business rules
- Backend domain rules should not depend directly on frontend behavior
- PostgreSQL should persist domain state without containing the main application logic
- Public demo interactions must remain controlled and isolated
- Infrastructure should remain proportional to the actual requirements of the MVP

Components such as caches, queues, background workers, WebSockets or additional services may be introduced later if the product creates a concrete need for them.

They are not required by the initial architecture.
