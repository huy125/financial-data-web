# Entity Relationship 🔗
### Description

A user can request an analysis for many stocks. Each analysis corresponds to a particular stock for a particular user.
A stock can have many metrics associated with it (e.g., P/E ratio, EPS, etc.).
A metric can be shared across multiple stocks.
A stock can have many analyses (for different users or at different times).
An analysis results in a recommendation for a stock.

```mermaid
erDiagram
    USER {
        int id PK
        string firstname
        string lastname
        date created_at
        date updated_at
    }

    STOCK {
        int id PK
        string symbol
        string company
    }

    METRIC {
        int id PK
        string name
        string description
    }

    STOCK_METRIC {
        int stock_id PK, FK
        int metric_id PK, FK
        float value
        date recorded_at
    }

    ANALYSIS {
        int id PK
        int user_id FK
        int stock_id FK
        int score
        date created_at
    }

    RECOMMENDATION {
        int id PK
        int analysis_id FK
        string action           "The action that user should take Buy, Hold Or Sell"
        float confidence_level  "A confidence score of the recommendation - in percentage"
        string reason           "A brief explanation of why the recommendation was made"
        date created_at
    }

    STOCK ||--o{ STOCK_METRIC : "contains"
    METRIC ||--o{ STOCK_METRIC : "be applied"
    USER ||--o{ ANALYSIS : "requests"
    STOCK ||--o{ ANALYSIS : "has"
    ANALYSIS ||--|| RECOMMENDATION : "concludes"
```
