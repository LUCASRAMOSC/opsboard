# OpsBoard — Demo Experience

## 1. Demo purpose

The OpsBoard public demo should allow visitors to understand the product's core idea without requiring configuration or prior knowledge of observability tools.

The demo uses the fictional **ShopFlow** workspace and should demonstrate how technical service degradation can affect user-facing Business Journeys.

The experience should help the visitor answer three questions:

1. What changed?
2. What is affected?
3. How important is the impact?

The demo is not intended to reproduce a real production monitoring environment.

Instead, it should provide a controlled and interactive scenario that demonstrates the OpsBoard domain and Impact Analysis capabilities.

## 2. Initial demo state

When the visitor opens the demo, the **ShopFlow** workspace should already contain representative sample data.

The initial environment includes Services such as:

- Web Frontend
- Auth API
- Orders API
- Payments API
- PostgreSQL

It also includes Business Journeys such as:

- Customer Login
- Checkout
- Order Tracking

The initial state should be mostly healthy.

This makes it easier for the visitor to understand the difference when an operational scenario is triggered.

The dashboard should provide an immediate overview of:

- Business Journey health
- Service health
- Active Incidents
- Recent Deployments
- Current operational impact

The visitor should be able to understand the general health of ShopFlow before interacting with any simulation.

## 3. Main dashboard experience

The main dashboard should communicate the current operational state of the ShopFlow workspace without requiring the visitor to navigate through multiple screens.

The first view should prioritize Business Journeys and operational impact instead of infrastructure metrics.

The dashboard should make it easy to identify:

- Healthy Business Journeys
- Degraded Business Journeys
- Down Business Journeys
- Affected Services
- Active Incidents
- Recent Deployments
- Current Impact Scores

Business Journeys should be visually more prominent than individual Services because they represent the user-facing impact that OpsBoard is designed to explain.

A visitor should be able to select a Business Journey and understand:

- Its current operational condition
- Its criticality
- Which Services it depends on
- Which dependency is currently causing degradation
- The current Impact Score
- Any active Incident related to the affected Services
- Recent Deployments that may provide operational context

The dashboard should avoid presenting low-level infrastructure metrics such as CPU, memory or disk usage.

The goal is to explain operational impact, not to reproduce a traditional monitoring dashboard.

## 4. Incident simulation flow

The public demo should allow the visitor to trigger predefined operational scenarios from the ShopFlow environment.

The simulation experience should be simple and intentional.

A visitor should be able to:

1. Choose a predefined scenario.
2. Start the simulation.
3. Observe the affected Service change state.
4. See the impact propagate to related Business Journeys.
5. View the updated Impact Score and severity.
6. Inspect possible operational context such as a recent Deployment.

Initial scenarios should include:

- Payments API high latency
- Payments API unavailable
- Auth API degradation

For example, when the visitor starts the **Payments API high latency** scenario, OpsBoard should create a new Health Event representing the degradation.

The expected flow is:

```text
Payments API
    ↓
DEGRADED
    ↓
Checkout
    ↓
HIGH IMPACT
    ↓
Possible deployment correlation
```

The simulation should make the relationship between technical degradation and Business Journey impact immediately understandable.

The visitor should not need to manually create Services, Health Events or Incidents before experiencing the core product behavior.

## 5. Impact visualization

After a degradation is triggered, the interface should make the resulting business impact clear.

The visitor should be able to understand:

- Which Service is degraded or unavailable
- Which Business Journeys are affected
- The Impact Score
- The impact severity
- Why the affected Journey received that impact level

The Impact Score should be presented as an explainable result rather than an isolated number.

The interface should prioritize clarity over technical detail.

## 6. Deployment correlation

When an affected Service has a recent Deployment, OpsBoard should provide that information as operational context.

The visitor should be able to see:

- Service
- Version
- Commit identifier
- Deployment timestamp
- Time between the Deployment and the beginning of the degradation

When the configured correlation rules are satisfied, the interface should display:

**Possible deployment correlation**

This information should help the visitor understand what changed before the problem appeared.

OpsBoard must never present this correlation as proof that the Deployment caused the degradation.

## 7. Demo reset and isolation

The visitor should be able to restore the ShopFlow demo to its initial state at any time.

Resetting the demo should:

- Restore the original Service health state
- Remove Health Events and Incidents created by simulations
- Restore the initial Health Events required by the demonstration
- Restore the original operational impact

Public demo interactions should be isolated so that one visitor cannot affect another visitor's experience.

The base ShopFlow dataset should remain unchanged.

## 8. Interaction boundaries

The public demo should remain controlled and safe.

Visitors may interact with predefined ShopFlow scenarios and explore the resulting operational state.

The public demo will not allow visitors to:

- Monitor arbitrary external URLs
- Register external network targets
- Modify the base ShopFlow dataset permanently
- Access another visitor's demo environment
- Execute unrestricted monitoring requests

These boundaries keep the demo focused on demonstrating OpsBoard's core product behavior while reducing unnecessary security and abuse risks.
