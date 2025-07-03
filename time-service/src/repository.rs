use diesel::{PgConnection, prelude::*, QueryResult};
use diesel::r2d2::{Pool, ConnectionManager, PooledConnection};
use uuid::Uuid;
use chrono::NaiveDateTime;
use crate::{
    models::{TimeEntry, CreateTimeEntry, TimeEntryResponse, UpdateTimeEntry}, 
    schema::time_entries, 
    errors::ServiceError
};

#[derive(Clone)]
pub struct TimeEntryRepository {
    pool: Pool<ConnectionManager<PgConnection>>,
}

impl TimeEntryRepository {
    pub fn new(database_url: &str) -> Self {
        let manager = ConnectionManager::<PgConnection>::new(database_url);
        let pool = Pool::builder()
            .build(manager)
            .expect("Failed to create pool.");
        TimeEntryRepository { pool }
    }

    fn conn(&self) -> Result<PooledConnection<ConnectionManager<PgConnection>>, ServiceError> {
        self.pool.get().map_err(|e| ServiceError::DatabaseError(e.to_string()))
    }

    pub fn create(&self, user_id: Uuid, entry: CreateTimeEntry) -> Result<TimeEntryResponse, ServiceError> {
        let conn = &mut self.conn()?;

        let new_entry = TimeEntry {
            id: Uuid::new_v4(),
            task_id: entry.task_id.unwrap(),
            user_id,
            start_time: chrono::Utc::now().naive_utc(),
            end_time: None,
            description: entry.description,
            created_at: chrono::Utc::now().naive_utc(),
            updated_at: chrono::Utc::now().naive_utc(),
        };

        diesel::insert_into(time_entries::table)
            .values(&new_entry)
            .execute(conn)?;

        Ok(TimeEntryResponse {
            id: new_entry.id,
            task_id: new_entry.task_id,
            user_id: new_entry.user_id,
            start_time: new_entry.start_time,
            end_time: new_entry.end_time,
            description: new_entry.description,
            duration: None,
        })
    }

    pub fn stop(&self, id: Uuid, user_id: Uuid) -> Result<TimeEntryResponse, ServiceError> {
        let conn = &mut self.conn()?;

        let entry: TimeEntry = time_entries::table
            .filter(time_entries::id.eq(id))
            .filter(time_entries::user_id.eq(user_id))
            .first(conn)?;

        let end_time = chrono::Utc::now().naive_utc();
        let duration = end_time.signed_duration_since(entry.start_time).num_seconds();

        diesel::update(time_entries::table)
            .filter(time_entries::id.eq(id))
            .set((
                time_entries::end_time.eq(end_time),
                time_entries::updated_at.eq(chrono::Utc::now().naive_utc()),
            ))
            .execute(conn)?;

        Ok(TimeEntryResponse {
            id: entry.id,
            task_id: entry.task_id,
            user_id: entry.user_id,
            start_time: entry.start_time,
            end_time: Some(end_time),
            description: entry.description,
            duration: Some(duration),
        })
    }

    pub fn update(&self, id: Uuid, user_id: Uuid, update: UpdateTimeEntry) -> Result<TimeEntryResponse, ServiceError> {
    let conn = &mut self.conn()?;

    let entry: TimeEntry = time_entries::table
        .filter(time_entries::id.eq(id))
        .filter(time_entries::user_id.eq(user_id))
        .first(conn)?;

    // Создаем копии значений перед использованием
    let description = update.description.clone();
    let end_time = update.end_time;

    diesel::update(time_entries::table)
        .filter(time_entries::id.eq(id))
        .set((
            time_entries::description.eq(&description),
            time_entries::end_time.eq(end_time),
            time_entries::updated_at.eq(chrono::Utc::now().naive_utc()),
        ))
        .execute(conn)?;

    let duration = match end_time {
        Some(end_time) => Some(end_time.signed_duration_since(entry.start_time).num_seconds()),
        None => None,
    };

    Ok(TimeEntryResponse {
        id: entry.id,
        task_id: entry.task_id,
        user_id: entry.user_id,
        start_time: entry.start_time,
        end_time,
        description,
        duration,
    })
    }

    pub fn list_by_user(&self, user_id: Uuid) -> Result<Vec<TimeEntryResponse>, ServiceError> {
        let conn = &mut self.conn()?;

        let entries = time_entries::table
            .filter(time_entries::user_id.eq(user_id))
            .order(time_entries::start_time.desc())
            .load::<TimeEntry>(conn)?;

        Ok(entries.into_iter().map(|entry| {
            let duration = entry.end_time.map(|end_time| {
                end_time.signed_duration_since(entry.start_time).num_seconds()
            });

            TimeEntryResponse {
                id: entry.id,
                task_id: entry.task_id,
                user_id: entry.user_id,
                start_time: entry.start_time,
                end_time: entry.end_time,
                description: entry.description,
                duration,
            }
        }).collect())
    }

    pub fn list_by_task(&self, user_id: Uuid, task_id: Uuid) -> Result<Vec<TimeEntryResponse>, ServiceError> {
        let conn = &mut self.conn()?;

        let entries = time_entries::table
            .filter(time_entries::user_id.eq(user_id))
            .filter(time_entries::task_id.eq(task_id))
            .order(time_entries::start_time.desc())
            .load::<TimeEntry>(conn)?;

        Ok(entries.into_iter().map(|entry| {
            let duration = entry.end_time.map(|end_time| {
                end_time.signed_duration_since(entry.start_time).num_seconds()
            });

            TimeEntryResponse {
                id: entry.id,
                task_id: entry.task_id,
                user_id: entry.user_id,
                start_time: entry.start_time,
                end_time: entry.end_time,
                description: entry.description,
                duration,
            }
        }).collect())
    }

    pub fn get_active(&self, user_id: Uuid) -> Result<Option<TimeEntryResponse>, ServiceError> {
        let conn = &mut self.conn()?;

        let entry = time_entries::table
            .filter(time_entries::user_id.eq(user_id))
            .filter(time_entries::end_time.is_null())
            .order(time_entries::start_time.desc())
            .first::<TimeEntry>(conn)
            .optional()?;

        Ok(entry.map(|entry| TimeEntryResponse {
            id: entry.id,
            task_id: entry.task_id,
            user_id: entry.user_id,
            start_time: entry.start_time,
            end_time: entry.end_time,
            description: entry.description,
            duration: None,
        }))
    }

    pub fn delete(&self, id: Uuid, user_id: Uuid) -> Result<(), ServiceError> {
        let conn = &mut self.conn()?;

        diesel::delete(time_entries::table)
            .filter(time_entries::id.eq(id))
            .filter(time_entries::user_id.eq(user_id))
            .execute(conn)?;

        Ok(())
    }
}