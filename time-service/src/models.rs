use chrono::{NaiveDateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use validator::Validate;
use diesel::{Queryable, Insertable};

#[derive(Debug, Clone, Serialize, Deserialize, Queryable, Insertable)]
#[diesel(table_name = crate::schema::time_entries)]
pub struct TimeEntry {
    pub id: Uuid,
    pub task_id: Uuid,
    pub user_id: Uuid,
    pub start_time: NaiveDateTime,
    pub end_time: Option<NaiveDateTime>,
    pub description: Option<String>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct CreateTimeEntry {
    #[validate(required)]
    pub task_id: Option<Uuid>,
    #[validate(length(max = 500))]
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct UpdateTimeEntry {
    #[validate(length(max = 500))]
    pub description: Option<String>,
    pub end_time: Option<NaiveDateTime>,
}

#[derive(Debug, Clone, Serialize)]
pub struct TimeEntryResponse {
    pub id: Uuid,
    pub task_id: Uuid,
    pub user_id: Uuid,
    pub start_time: NaiveDateTime,
    pub end_time: Option<NaiveDateTime>,
    pub description: Option<String>,
    pub duration: Option<i64>, // in seconds
}