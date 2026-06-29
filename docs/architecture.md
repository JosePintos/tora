# Architecture

## High-Level Diagram

![high-level-arch.drawio.png](Architecture/high-level-arch.drawio.png)

Notice there's only **one deployable backend**.

GraphQL is just another adapter.

## Hexagonal Architecture

![hexagonal.png](Architecture/hexagonal.png)

No business logic should know GraphQL exists.

No business logic should know PostgreSQL exists.