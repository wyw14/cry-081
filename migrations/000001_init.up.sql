CREATE DOMAIN entity_id AS text
    NOT NULL
    CONSTRAINT entity_id_not_blank CHECK (length(btrim(VALUE)) > 0);
CREATE DOMAIN required_text AS text
    NOT NULL
    CONSTRAINT required_text_not_blank CHECK (length(btrim(VALUE)) > 0);
CREATE DOMAIN entity_version AS bigint
    NOT NULL
    CONSTRAINT entity_version_positive CHECK (VALUE > 0);
CREATE DOMAIN positive_integer AS integer
    NOT NULL
    CONSTRAINT positive_integer_value CHECK (VALUE > 0);
CREATE DOMAIN nonnegative_integer AS integer
    NOT NULL
    CONSTRAINT nonnegative_integer_value CHECK (VALUE >= 0);
CREATE DOMAIN utc_instant AS timestamp with time zone NOT NULL;

CREATE TYPE manuscript_status AS ENUM (
    'draft', 'submitted', 'under_initial_review', 'revision_needed',
    'under_re_review', 'under_final_review', 'rejected', 'accepted',
    'scheduled', 'published', 'withdrawn'
);
CREATE TYPE assignment_status AS ENUM ('offered', 'accepted', 'declined', 'completed', 'cancelled');
CREATE TYPE issue_status AS ENUM ('planning', 'locked', 'released');
CREATE TYPE article_status AS ENUM ('scheduled', 'published', 'withdrawn');

CREATE DOMAIN manuscript_state AS manuscript_status NOT NULL;
CREATE DOMAIN assignment_state AS assignment_status NOT NULL;
CREATE DOMAIN issue_state AS issue_status NOT NULL;
CREATE DOMAIN article_state AS article_status NOT NULL;

CREATE TABLE users (
    id entity_id,
    email required_text,
    display_name required_text,
    password_hash bytea NOT NULL,
    roles jsonb NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at utc_instant,
    version entity_version,
    CONSTRAINT users_pk PRIMARY KEY (id),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_roles_array CHECK (jsonb_typeof(roles) = 'array')
);

CREATE TABLE refresh_tokens (
    id entity_id,
    user_id entity_id,
    digest required_text,
    issued_at utc_instant,
    expires_at utc_instant,
    revoked_at timestamp with time zone,
    version entity_version,
    CONSTRAINT refresh_tokens_pk PRIMARY KEY (id),
    CONSTRAINT refresh_tokens_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT refresh_tokens_digest_unique UNIQUE (digest),
    CONSTRAINT refresh_tokens_time_order CHECK (expires_at > issued_at)
);

CREATE TABLE manuscripts (
    id entity_id,
    author_id entity_id,
    section_id entity_id,
    status manuscript_state,
    active_version entity_version,
    fingerprint required_text,
    aggregate jsonb NOT NULL,
    version entity_version,
    created_at utc_instant,
    updated_at utc_instant,
    CONSTRAINT manuscripts_pk PRIMARY KEY (id),
    CONSTRAINT manuscripts_author_fk FOREIGN KEY (author_id) REFERENCES users (id)
);

CREATE INDEX manuscripts_author_updated_idx ON manuscripts (author_id, updated_at DESC);
CREATE INDEX manuscripts_section_status_idx ON manuscripts (section_id, status, updated_at DESC);
CREATE INDEX manuscripts_fingerprint_idx ON manuscripts (fingerprint);

CREATE TABLE idempotency_records (
    scope required_text,
    key required_text,
    request_digest required_text,
    response_body bytea NOT NULL,
    status_code positive_integer,
    expires_at utc_instant,
    CONSTRAINT idempotency_records_pk PRIMARY KEY (scope, key),
    CONSTRAINT idempotency_status_http CHECK (status_code BETWEEN 200 AND 599)
);

CREATE INDEX idempotency_records_expiry_idx ON idempotency_records (expires_at);

CREATE TABLE editorial_conflicts (
    editor_id entity_id,
    author_id entity_id,
    reason required_text,
    starts_at utc_instant,
    ends_at timestamp with time zone,
    CONSTRAINT editorial_conflicts_pk PRIMARY KEY (editor_id, author_id, starts_at),
    CONSTRAINT editorial_conflicts_editor_fk FOREIGN KEY (editor_id) REFERENCES users (id),
    CONSTRAINT editorial_conflicts_author_fk FOREIGN KEY (author_id) REFERENCES users (id),
    CONSTRAINT editorial_conflicts_distinct_people CHECK (editor_id <> author_id),
    CONSTRAINT editorial_conflicts_time_order CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX editorial_conflicts_active_idx ON editorial_conflicts (editor_id, author_id, ends_at);

CREATE TABLE review_assignments (
    id entity_id,
    manuscript_id entity_id,
    manuscript_version entity_version,
    reviewer_id entity_id,
    due_at utc_instant,
    status assignment_state,
    aggregate jsonb NOT NULL,
    version entity_version,
    CONSTRAINT review_assignments_pk PRIMARY KEY (id),
    CONSTRAINT review_assignments_manuscript_fk FOREIGN KEY (manuscript_id) REFERENCES manuscripts (id),
    CONSTRAINT review_assignments_reviewer_fk FOREIGN KEY (reviewer_id) REFERENCES users (id),
    CONSTRAINT review_assignments_one_reviewer UNIQUE (manuscript_id, manuscript_version, reviewer_id)
);

CREATE INDEX review_assignments_due_idx ON review_assignments (status, due_at);
CREATE INDEX review_assignments_manuscript_idx ON review_assignments (manuscript_id, manuscript_version);

CREATE TABLE review_reports (
    id entity_id,
    assignment_id entity_id,
    manuscript_id entity_id,
    manuscript_version entity_version,
    reviewer_id entity_id,
    aggregate jsonb NOT NULL,
    version entity_version,
    CONSTRAINT review_reports_pk PRIMARY KEY (id),
    CONSTRAINT review_reports_assignment_fk FOREIGN KEY (assignment_id) REFERENCES review_assignments (id),
    CONSTRAINT review_reports_manuscript_fk FOREIGN KEY (manuscript_id) REFERENCES manuscripts (id),
    CONSTRAINT review_reports_reviewer_fk FOREIGN KEY (reviewer_id) REFERENCES users (id),
    CONSTRAINT review_reports_assignment_unique UNIQUE (assignment_id)
);

CREATE INDEX review_reports_manuscript_idx ON review_reports (manuscript_id, manuscript_version);

CREATE TABLE editorial_decisions (
    id entity_id,
    manuscript_id entity_id,
    manuscript_version entity_version,
    aggregate jsonb NOT NULL,
    created_at utc_instant,
    CONSTRAINT editorial_decisions_pk PRIMARY KEY (id),
    CONSTRAINT editorial_decisions_manuscript_fk FOREIGN KEY (manuscript_id) REFERENCES manuscripts (id)
);

CREATE INDEX editorial_decisions_manuscript_idx ON editorial_decisions (manuscript_id, created_at);

CREATE TABLE issues (
    id entity_id,
    volume positive_integer,
    number positive_integer,
    status issue_state,
    release_at utc_instant,
    aggregate jsonb NOT NULL,
    version entity_version,
    CONSTRAINT issues_pk PRIMARY KEY (id),
    CONSTRAINT issues_volume_number_unique UNIQUE (volume, number)
);

CREATE TABLE articles (
    id entity_id,
    manuscript_id entity_id,
    issue_id entity_id,
    section_id entity_id,
    title required_text,
    status article_state,
    published_at timestamp with time zone,
    aggregate jsonb NOT NULL,
    version entity_version,
    CONSTRAINT articles_pk PRIMARY KEY (id),
    CONSTRAINT articles_manuscript_fk FOREIGN KEY (manuscript_id) REFERENCES manuscripts (id),
    CONSTRAINT articles_issue_fk FOREIGN KEY (issue_id) REFERENCES issues (id),
    CONSTRAINT articles_manuscript_unique UNIQUE (manuscript_id)
);

CREATE INDEX articles_public_search_idx ON articles (status, published_at DESC);
CREATE INDEX articles_section_idx ON articles (section_id, published_at DESC);

CREATE TABLE audit_events (
    id entity_id,
    actor_id entity_id,
    source required_text,
    action required_text,
    resource required_text,
    resource_id entity_id,
    before_value jsonb NOT NULL,
    after_value jsonb NOT NULL,
    reason required_text,
    created_at utc_instant,
    CONSTRAINT audit_events_pk PRIMARY KEY (id)
);

CREATE INDEX audit_events_resource_idx ON audit_events (resource, resource_id, created_at DESC);
CREATE INDEX audit_events_actor_idx ON audit_events (actor_id, created_at DESC);

CREATE TABLE outbox_messages (
    id entity_id,
    topic required_text,
    message_key entity_id,
    payload bytea NOT NULL,
    attempts nonnegative_integer DEFAULT 0,
    maximum_attempts positive_integer,
    available_at utc_instant,
    processed_at timestamp with time zone,
    dead_at timestamp with time zone,
    last_error text DEFAULT '' NOT NULL,
    CONSTRAINT outbox_messages_pk PRIMARY KEY (id)
);

CREATE INDEX outbox_messages_available_idx
    ON outbox_messages (available_at)
    WHERE processed_at IS NULL AND dead_at IS NULL;
