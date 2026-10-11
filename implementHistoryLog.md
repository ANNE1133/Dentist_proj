# Task: Implement Camera Detection History Logging and Camera Log Page

The current system already has a working camera/object detection flow.

However, **detected people are NOT currently being saved into `camera_history_logs`**.

The task is to implement the complete flow:

```text
Camera
  ↓
Object Detection
  ↓
Face Recognition (if already implemented)
  ↓
Create Camera History Log
  ↓
PostgreSQL: camera_history_logs
  ↓
Camera Log Page
  ↓
Query by time + pagination
```

Do NOT implement the 3D chair integration yet.

Do NOT implement adjustment calculation yet.

---

# 1. Existing Database

The camera history table already exists:

```sql
CREATE TABLE camera_history_logs (
    id SERIAL PRIMARY KEY,
    dentist_id VARCHAR(100) NOT NULL,
    body_composition TEXT,
    adjustment_items TEXT[],
    adjustment_status VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

There is also an existing master table:

```sql
CREATE TABLE adjustable_items (
    id SERIAL PRIMARY KEY,
    adjust_item_id VARCHAR(100) UNIQUE NOT NULL,
    adjust_name VARCHAR(255) NOT NULL,
    data_used TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

`adjustable_items` is currently only a master/reference table.

The actual adjustment calculation has NOT been implemented.

---

# 2. IMPORTANT: Inspect Existing Camera Flow First

Before writing code, inspect the existing implementation.

Find:

- Camera connection
- Camera stream
- Object detection
- Person detection
- Face recognition
- Dentist identification
- Body composition detection
- Existing backend API calls
- Existing database service/repository
- Existing Prisma/SQL/ORM configuration

Determine the exact point where the system knows:

```text
Person detected
```

and, if available:

```text
Person recognized as Dentist
```

The new logging logic should be inserted into the existing flow rather than creating a separate detection system.

Do NOT rewrite the existing camera/object detection implementation.

---

# 3. Create Camera History Log After Detection

When the camera successfully detects a person, create a record in:

```text
camera_history_logs
```

The basic flow should be:

```text
Camera detects person
       ↓
Determine identity if face recognition is available
       ↓
Prepare camera history data
       ↓
INSERT camera_history_logs
```

Example SQL:

```sql
INSERT INTO camera_history_logs (
    dentist_id,
    body_composition,
    adjustment_items,
    adjustment_status
)
VALUES (
    :dentistId,
    :bodyComposition,
    :adjustmentItems,
    :adjustmentStatus
);
```

`created_at` should use the database default.

Do not manually generate `created_at` unless the existing architecture requires it.

---

# 4. Recognized Dentist

If face recognition identifies the person as a dentist:

Example:

```text
dentist_id = "D001"
```

Create:

```json
{
    "dentistId": "D001",
    "bodyComposition": "...",
    "adjustmentItems": null,
    "adjustmentStatus": "pending"
}
```

IMPORTANT:

`adjustment_items` must NOT be invented.

There is currently no adjustment calculation logic.

Therefore, use:

```text
NULL
```

or the existing project's appropriate empty-value convention.

Do NOT generate:

```text
seat_height
backrest_angle
headrest_position
```

just for testing.

---

# 5. Unknown Person

If object detection detects a person but face recognition does not identify them:

Use the existing project convention for an unknown person.

For example, if the project already uses:

```text
UNKNOWN
```

then use:

```text
dentist_id = "UNKNOWN"
```

Do NOT introduce a new convention if the existing code already has one.

The record should still be saved into:

```text
camera_history_logs
```

Example:

```json
{
    "dentistId": "UNKNOWN",
    "bodyComposition": "...",
    "adjustmentItems": null,
    "adjustmentStatus": "not_applicable"
}
```

Unknown people should NOT receive personalized adjustment data.

---

# 6. Avoid Duplicate Logs

This is important.

A camera may continuously detect the same person across many frames.

Do NOT insert a new database row for every video frame.

For example, this is WRONG:

```text
Frame 1 → INSERT
Frame 2 → INSERT
Frame 3 → INSERT
Frame 4 → INSERT
...
```

Instead, implement a reasonable logging mechanism so that one continuous detection does not generate hundreds/thousands of duplicate records.

First inspect the existing detection/recognition flow to determine whether it already has:

- detection events
- recognition events
- tracking IDs
- cooldown/debounce
- state changes

Prefer using the existing mechanism.

If there is no existing mechanism, implement a simple configurable cooldown/event-based approach.

Example concept:

```text
Person detected
      ↓
Create log
      ↓
Ignore same detection for N seconds
      ↓
Allow another log after cooldown
```

Do NOT hardcode an arbitrary value without checking the existing application's requirements. If a default is necessary, make it configurable.

---

# 7. Backend API

Implement the backend functionality for creating camera history logs.

If the existing architecture already has an appropriate endpoint/service, extend it instead of creating a duplicate.

If no suitable implementation exists, create a service such as:

```text
CameraHistoryService
```

with a function conceptually similar to:

```text
createCameraHistoryLog(...)
```

The service should:

1. Validate input.
2. Insert into `camera_history_logs`.
3. Return the created record.
4. Handle database errors properly.

---

# 8. Camera Log Query API

Implement an API for retrieving camera history.

Example:

```http
GET /camera-history
```

Support:

```text
page
limit
startTime
endTime
```

Example:

```http
GET /camera-history?page=1&limit=20&startTime=2026-10-07T00:00:00&endTime=2026-10-07T23:59:59
```

Query:

```sql
SELECT
    id,
    dentist_id,
    body_composition,
    adjustment_items,
    adjustment_status,
    created_at
FROM camera_history_logs
WHERE created_at >= :startTime
  AND created_at <= :endTime
ORDER BY created_at DESC
LIMIT :limit
OFFSET :offset;
```

Calculate:

```text
offset = (page - 1) * limit
```

Also return total count for pagination.

---

# 9. Camera Log Page

Create/implement the Camera Log Page.

The page must retrieve data from the backend rather than local state/localStorage.

Display:

```text
Time
Person / Dentist ID
Recognition Status
Body Composition
Adjustment Items
Adjustment Status
```

Example recognized dentist:

```text
----------------------------------------------------
10:30:25

Dentist: D001
Status: Recognized Dentist

Body Composition:
...

Adjustment Items:
Not determined yet

Adjustment Status:
Pending
----------------------------------------------------
```

Example unknown:

```text
----------------------------------------------------
10:25:10

Person: Unknown
Status: Unknown Person

Body Composition:
...

Adjustment Items:
None

Adjustment Status:
N/A
----------------------------------------------------
```

---

# 10. Time Filtering

The user must be able to select:

```text
Start Date/Time
End Date/Time
```

Then request:

```text
GET /camera-history
```

with the selected range.

Example:

```text
From:
2026-10-07 09:00

To:
2026-10-07 12:00
```

The backend must perform the time filtering in PostgreSQL.

Do NOT fetch all records and filter them only in the frontend.

---

# 11. Pagination

Implement pagination.

Example:

```text
Page 1
20 records

< Previous   1  2  3  4  5   Next >
```

The frontend should request only the required page.

Do not load the entire history table.

---

# 12. Adjustment Items

For now:

```text
adjustment_items = NULL
```

or the existing project's appropriate empty representation.

The system must NOT determine personalized adjustment values yet.

The existing:

```text
adjustable_items
```

table is only a reference/master table.

It means:

```text
"What adjustment capabilities does the chair support?"
```

It does NOT currently mean:

```text
"What should this specific dentist's chair position be?"
```

That logic will be implemented later.

---

# 13. Future 3D Integration

After the camera log is successfully created, the system will eventually trigger another service.

Future flow:

```text
Camera Detection
       ↓
Camera History Log
       ↓
Adjustment Calculation
       ↓
3D Chair Service
```

However, **do NOT implement the actual 3D communication now**.

Prepare a clean service boundary, for example:

```text
createCameraHistoryLog()
        ↓
[future]
AdjustmentService
        ↓
[future]
3D Chair Service
```

The future payload may eventually look like:

```json
{
    "historyLogId": 123,
    "dentistId": "D001",
    "bodyComposition": "...",
    "adjustmentItems": [
        "seat_height",
        "backrest_angle"
    ]
}
```

But currently `adjustmentItems` should remain empty/null because the calculation logic does not exist.

---

# 14. Database Index

Because the Camera Log Page will query by time, check whether an index exists for:

```text
created_at
```

If no suitable index exists, add:

```sql
CREATE INDEX IF NOT EXISTS idx_camera_history_logs_created_at
ON camera_history_logs(created_at);
```

Do not create duplicate indexes.

---

# 15. Error Handling

Handle:

### Database failure

Return an appropriate server error.

### Invalid time range

For example:

```text
startTime > endTime
```

Return:

```text
400 Bad Request
```

### No records

Return an empty result:

```json
{
    "data": [],
    "page": 1,
    "limit": 20,
    "total": 0,
    "totalPages": 0
}
```

Frontend should show:

```text
No camera history found.
```

---

# 16. Required Final Flow

After implementation, the expected current system is:

```text
┌──────────────┐
│    Camera    │
└──────┬───────┘
       ↓
┌──────────────┐
│    Object    │
│  Detection   │
└──────┬───────┘
       ↓
┌──────────────┐
│    Face      │
│ Recognition  │
└──────┬───────┘
       ↓
┌─────────────────────┐
│ Dentist / Unknown   │
└──────────┬──────────┘
           ↓
┌─────────────────────────┐
│ Create Camera History   │
│ Log                     │
└──────────┬──────────────┘
           ↓
┌─────────────────────────┐
│ PostgreSQL              │
│ camera_history_logs     │
└──────────┬──────────────┘
           ↓
┌─────────────────────────┐
│ Camera Log API           │
│ time filter + pagination │
└──────────┬──────────────┘
           ↓
┌─────────────────────────┐
│ Camera Log Page          │
└─────────────────────────┘
```

Future:

```text
camera_history_logs
       ↓
Adjustment Calculation
       ↓
adjustment_items
       ↓
3D Chair Service
```

---

# 17. Acceptance Criteria

The implementation is complete when:

- [ ] Existing camera detection continues to work.
- [ ] A detected person creates a row in `camera_history_logs`.
- [ ] Recognized dentist stores the correct `dentist_id`.
- [ ] Unknown person is logged using the project's existing unknown convention.
- [ ] `body_composition` is stored when available.
- [ ] `adjustment_items` is NOT artificially generated.
- [ ] `adjustment_status` follows the current application's convention.
- [ ] Duplicate logs caused by every camera frame are prevented.
- [ ] Camera Log API can query logs.
- [ ] Time filtering works in the database.
- [ ] Pagination works.
- [ ] Camera Log Page displays the records.
- [ ] Empty results are handled.
- [ ] Database errors are handled.
- [ ] `created_at` is used for chronological history.
- [ ] `adjustable_items` is treated only as a master/reference table.
- [ ] No adjustment calculation is implemented.
- [ ] No 3D integration is implemented.
- [ ] A clean boundary is prepared for future adjustment → 3D integration.

Finally, test the complete flow with an actual camera detection and verify that a row is inserted into:

```text
camera_history_logs
```

Then verify that the same record appears correctly on the Camera Log Page.