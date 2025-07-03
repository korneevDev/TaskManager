// @generated automatically by Diesel CLI.

diesel::table! {
    tasks (id) {
        id -> Int8,
        title -> Nullable<Text>,
        description -> Nullable<Text>,
        status -> Nullable<Text>,
        created_at -> Nullable<Timestamptz>,
        updated_at -> Nullable<Timestamptz>,
        user_id -> Nullable<Int8>,
    }
}

diesel::table! {
    time_entries (id) {
        id -> Uuid,
        task_id -> Uuid,
        user_id -> Uuid,
        start_time -> Timestamp,
        end_time -> Nullable<Timestamp>,
        description -> Nullable<Text>,
        created_at -> Timestamp,
        updated_at -> Timestamp,
    }
}

diesel::table! {
    users (id) {
        id -> Int4,
        username -> Nullable<Text>,
        password -> Text,
        created_at -> Nullable<Timestamptz>,
        updated_at -> Nullable<Timestamptz>,
        deleted_at -> Nullable<Timestamptz>,
        refresh_token -> Nullable<Text>,
    }
}

diesel::allow_tables_to_appear_in_same_query!(
    tasks,
    time_entries,
    users,
);
