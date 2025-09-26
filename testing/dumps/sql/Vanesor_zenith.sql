-- Downloaded from: https://github.com/Vanesor/zenith/blob/c6e7ebf9e6e71af74a410b62b344105c301743a2/schema_new.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.18 (Ubuntu 14.18-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 14.18 (Ubuntu 14.18-0ubuntu0.22.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: generate_task_key(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_task_key() RETURNS trigger
    LANGUAGE plpgsql
    AS $_$
DECLARE
    project_key_val text;
    next_number integer;
BEGIN
    -- Get project key
    SELECT project_key INTO project_key_val 
    FROM projects 
    WHERE id = NEW.project_id;
    
    -- Get next task number for this project
    SELECT COALESCE(MAX(CAST(SUBSTRING(task_key FROM '[0-9]+$') AS integer)), 0) + 1
    INTO next_number
    FROM tasks 
    WHERE project_id = NEW.project_id;
    
    -- Generate the task key
    NEW.task_key := project_key_val || '-' || next_number;
    
    RETURN NEW;
END;
$_$;


--
-- Name: update_post_search_vector(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_post_search_vector() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', 
        COALESCE(NEW.title, '') || ' ' || 
        COALESCE(NEW.content, '') || ' ' || 
        COALESCE(NEW.excerpt, '') || ' ' ||
        COALESCE(array_to_string(NEW.tags, ' '), '')
    );
    RETURN NEW;
END;
$$;


--
-- Name: update_project_progress(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_project_progress() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_tasks_count integer;
    completed_tasks_count integer;
    progress_pct numeric;
BEGIN
    -- Count total and completed tasks for the project
    SELECT 
        COUNT(*) as total,
        COUNT(*) FILTER (WHERE is_completed = true) as completed
    INTO total_tasks_count, completed_tasks_count
    FROM tasks 
    WHERE project_id = COALESCE(NEW.project_id, OLD.project_id);
    
    -- Calculate progress percentage
    IF total_tasks_count > 0 THEN
        progress_pct := (completed_tasks_count::numeric / total_tasks_count::numeric) * 100;
    ELSE
        progress_pct := 0;
    END IF;
    
    -- Update project
    UPDATE projects 
    SET 
        total_tasks = total_tasks_count,
        completed_tasks = completed_tasks_count,
        progress_percentage = progress_pct,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.project_id, OLD.project_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ai_assignment_generations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_assignment_generations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    template_id uuid,
    generated_assignment_id uuid,
    source_file_url text NOT NULL,
    generation_prompt text,
    ai_model_used character varying,
    generation_status character varying DEFAULT 'pending'::character varying,
    questions_extracted integer DEFAULT 0,
    questions_created integer DEFAULT 0,
    processing_log jsonb DEFAULT '[]'::jsonb,
    error_details text,
    generated_by uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    completed_at timestamp with time zone
);


--
-- Name: announcements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.announcements (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying NOT NULL,
    content text NOT NULL,
    author_id uuid,
    club_id character varying,
    priority character varying DEFAULT 'normal'::character varying,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: assignment_attempts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignment_attempts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    assignment_id uuid NOT NULL,
    user_id uuid NOT NULL,
    attempt_number integer DEFAULT 1 NOT NULL,
    start_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    end_time timestamp with time zone,
    time_spent integer DEFAULT 0,
    score integer DEFAULT 0,
    max_score integer DEFAULT 0,
    percentage numeric DEFAULT 0,
    is_passing boolean DEFAULT false,
    answers jsonb DEFAULT '{}'::jsonb,
    graded_answers jsonb DEFAULT '{}'::jsonb,
    violations jsonb DEFAULT '[]'::jsonb,
    status character varying DEFAULT 'in_progress'::character varying,
    submitted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    is_fullscreen boolean DEFAULT false,
    auto_save_data jsonb DEFAULT '{}'::jsonb,
    window_violations integer DEFAULT 0,
    last_auto_save timestamp with time zone,
    browser_info jsonb DEFAULT '{}'::jsonb
);


--
-- Name: assignment_audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignment_audit_log (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    assignment_id uuid NOT NULL,
    user_id uuid NOT NULL,
    attempt_id uuid,
    action character varying NOT NULL,
    details jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: assignment_questions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignment_questions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    assignment_id uuid NOT NULL,
    question_text text NOT NULL,
    question_type character varying NOT NULL,
    marks integer DEFAULT 1 NOT NULL,
    time_limit integer,
    code_language character varying,
    code_template text,
    test_cases jsonb,
    expected_output text,
    solution text,
    ordering integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    type character varying,
    title character varying,
    description text,
    options jsonb,
    correct_answer jsonb,
    points integer DEFAULT 1,
    question_order integer DEFAULT 0,
    starter_code text,
    integer_min numeric,
    integer_max numeric,
    integer_step numeric DEFAULT 1,
    explanation text,
    allowed_languages jsonb DEFAULT '[]'::jsonb,
    allow_any_language boolean DEFAULT false,
    question_image_url text,
    question_image_alt text,
    question_images jsonb DEFAULT '[]'::jsonb,
    answer_images jsonb DEFAULT '[]'::jsonb,
    CONSTRAINT assignment_questions_question_type_check CHECK (((question_type)::text = ANY (ARRAY[('single_choice'::character varying)::text, ('multiple_choice'::character varying)::text, ('multi_select'::character varying)::text, ('coding'::character varying)::text, ('essay'::character varying)::text, ('true_false'::character varying)::text, ('integer'::character varying)::text])))
);


--
-- Name: assignment_submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignment_submissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    assignment_id uuid,
    user_id uuid,
    submission_text text,
    file_url text,
    submitted_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    status character varying DEFAULT 'submitted'::character varying,
    grade integer,
    feedback text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    violation_count integer DEFAULT 0,
    time_spent integer,
    auto_submitted boolean DEFAULT false,
    ip_address character varying,
    user_agent text,
    total_score integer DEFAULT 0
);


--
-- Name: assignment_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignment_templates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying NOT NULL,
    description text,
    template_file_url text NOT NULL,
    template_type character varying NOT NULL,
    category character varying,
    subject character varying,
    difficulty_level character varying,
    estimated_questions integer,
    created_by uuid,
    is_active boolean DEFAULT true,
    usage_count integer DEFAULT 0,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: assignment_violations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignment_violations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    submission_id uuid NOT NULL,
    violation_type character varying NOT NULL,
    occurred_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    details jsonb
);


--
-- Name: assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assignments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying NOT NULL,
    description text NOT NULL,
    club_id character varying,
    created_by uuid,
    due_date timestamp with time zone NOT NULL,
    max_points integer DEFAULT 100,
    instructions text,
    status character varying DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    assignment_type character varying DEFAULT 'regular'::character varying,
    target_audience character varying DEFAULT 'club'::character varying,
    target_clubs character varying[] DEFAULT '{}'::character varying[],
    time_limit integer,
    allow_navigation boolean DEFAULT true,
    passing_score integer DEFAULT 60,
    is_proctored boolean DEFAULT false,
    shuffle_questions boolean DEFAULT false,
    allow_calculator boolean DEFAULT true,
    show_results boolean DEFAULT true,
    allow_review boolean DEFAULT true,
    shuffle_options boolean DEFAULT false,
    max_attempts integer DEFAULT 1,
    is_published boolean DEFAULT false,
    coding_instructions text DEFAULT 'Write your code solution. Make sure to test your code thoroughly before submitting.'::text,
    objective_instructions text DEFAULT 'Choose the correct answer(s) for each question. For multi-select questions, you may choose multiple options.'::text,
    mixed_instructions text DEFAULT 'This assignment contains different types of questions. Read each question carefully and provide appropriate answers.'::text,
    essay_instructions text DEFAULT 'Provide detailed written responses to the essay questions. Ensure your answers are well-structured and comprehensive.'::text,
    require_fullscreen boolean DEFAULT false,
    auto_submit_on_violation boolean DEFAULT false,
    max_violations integer DEFAULT 3,
    code_editor_settings jsonb DEFAULT '{"theme": "vs-dark", "autoSave": true, "fontSize": 14, "wordWrap": true, "autoSaveInterval": 30000}'::jsonb,
    require_camera boolean DEFAULT false,
    require_microphone boolean DEFAULT false,
    require_face_verification boolean DEFAULT false,
    proctoring_settings jsonb DEFAULT '{}'::jsonb,
    start_date timestamp with time zone,
    start_time timestamp with time zone,
    CONSTRAINT assignments_assignment_type_check CHECK (((assignment_type)::text = ANY (ARRAY[('regular'::character varying)::text, ('objective'::character varying)::text, ('coding'::character varying)::text, ('essay'::character varying)::text]))),
    CONSTRAINT assignments_target_audience_check CHECK (((target_audience)::text = ANY (ARRAY[('club'::character varying)::text, ('all_clubs'::character varying)::text, ('specific_clubs'::character varying)::text])))
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    action character varying NOT NULL,
    resource_type character varying NOT NULL,
    resource_id uuid,
    old_values jsonb,
    new_values jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: chat_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_attachments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    message_id uuid,
    room_id uuid NOT NULL,
    filename character varying NOT NULL,
    original_filename character varying NOT NULL,
    file_path character varying NOT NULL,
    file_type character varying NOT NULL,
    file_size integer NOT NULL,
    mime_type character varying,
    encryption_key text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: chat_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_invitations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    room_id uuid NOT NULL,
    inviter_id uuid NOT NULL,
    invitee_email character varying NOT NULL,
    invitation_token character varying NOT NULL,
    message text,
    status character varying DEFAULT 'pending'::character varying,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    accepted_at timestamp with time zone
);


--
-- Name: chat_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    room_id uuid,
    user_id uuid,
    message text NOT NULL,
    message_type character varying DEFAULT 'text'::character varying,
    file_url text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    reply_to_message_id uuid,
    is_edited boolean DEFAULT false,
    reply_to uuid,
    sender_id uuid,
    content text,
    is_encrypted boolean DEFAULT false,
    updated_at timestamp with time zone,
    attachments jsonb DEFAULT '[]'::jsonb,
    message_images jsonb DEFAULT '[]'::jsonb,
    reactions jsonb DEFAULT '{}'::jsonb,
    thread_id uuid,
    edited_at timestamp with time zone,
    edited_by uuid,
    can_edit_until timestamp with time zone
);


--
-- Name: chat_room_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_room_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chat_room_id uuid,
    user_id uuid,
    joined_at timestamp with time zone DEFAULT now(),
    role character varying DEFAULT 'member'::character varying,
    user_email character varying
);


--
-- Name: chat_rooms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_rooms (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying NOT NULL,
    description text,
    club_id character varying,
    type character varying DEFAULT 'public'::character varying,
    created_by uuid,
    members uuid[] DEFAULT '{}'::uuid[],
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    room_type character varying DEFAULT 'public'::character varying,
    encryption_enabled boolean DEFAULT false,
    cover_image_url text,
    room_images jsonb DEFAULT '[]'::jsonb,
    room_settings jsonb DEFAULT '{}'::jsonb,
    profile_picture_url text,
    edited_at timestamp with time zone,
    edited_by uuid,
    CONSTRAINT chat_room_type_check CHECK (((type)::text = ANY ((ARRAY['public'::character varying, 'club'::character varying])::text[])))
);


--
-- Name: club_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.club_members (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    club_id uuid NOT NULL,
    is_leader boolean DEFAULT false,
    joined_at timestamp with time zone DEFAULT now()
);


--
-- Name: club_statistics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.club_statistics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: club_statistics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.club_statistics (
    id integer DEFAULT nextval('public.club_statistics_id_seq'::regclass) NOT NULL,
    club_id character varying,
    member_count integer DEFAULT 0,
    event_count integer DEFAULT 0,
    assignment_count integer DEFAULT 0,
    comment_count integer DEFAULT 0,
    total_engagement integer DEFAULT 0,
    average_engagement numeric DEFAULT 0,
    last_updated timestamp with time zone DEFAULT now()
);


--
-- Name: clubs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clubs (
    id character varying NOT NULL,
    name character varying NOT NULL,
    type character varying NOT NULL,
    description text NOT NULL,
    long_description text,
    icon character varying NOT NULL,
    color character varying NOT NULL,
    coordinator_id uuid,
    co_coordinator_id uuid,
    secretary_id uuid,
    media_id uuid,
    guidelines text,
    meeting_schedule jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    logo_url text,
    banner_image_url text,
    club_images jsonb DEFAULT '[]'::jsonb,
    member_count integer DEFAULT 0
);


--
-- Name: code_results; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.code_results (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    response_id uuid NOT NULL,
    test_case_index integer,
    passed boolean,
    stdout text,
    stderr text,
    execution_time integer,
    memory_used integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: coding_submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coding_submissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    question_response_id uuid NOT NULL,
    language character varying NOT NULL,
    code text NOT NULL,
    is_final boolean DEFAULT false,
    execution_result jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    post_id uuid,
    author_id uuid,
    content text NOT NULL,
    parent_id uuid,
    likes_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: committee_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.committee_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    committee_id uuid NOT NULL,
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    status character varying DEFAULT 'active'::character varying,
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    term_start timestamp with time zone,
    term_end timestamp with time zone,
    achievements jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: committee_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.committee_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    committee_id uuid NOT NULL,
    name character varying NOT NULL,
    description text,
    hierarchy integer DEFAULT 1 NOT NULL,
    permissions text[] DEFAULT '{}'::text[],
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: committees; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.committees (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description text,
    hierarchy_level integer DEFAULT 1 NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: discussion_replies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.discussion_replies (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    discussion_id uuid,
    author_id uuid,
    content text NOT NULL,
    parent_id uuid,
    likes_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: discussions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.discussions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying NOT NULL,
    description text,
    author_id uuid,
    club_id character varying,
    tags text[] DEFAULT '{}'::text[],
    is_locked boolean DEFAULT false,
    is_pinned boolean DEFAULT false,
    views_count integer DEFAULT 0,
    replies_count integer DEFAULT 0,
    last_activity timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: email_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    recipient character varying NOT NULL,
    subject character varying NOT NULL,
    content_preview text,
    status character varying DEFAULT 'sent'::character varying,
    message_id character varying,
    category character varying,
    related_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    sent_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    email_service character varying DEFAULT 'resend'::character varying,
    error_message text,
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: event_attendees; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_attendees (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    event_id uuid,
    user_id uuid,
    registered_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    attendance_status character varying DEFAULT 'registered'::character varying
);


--
-- Name: event_registrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_registrations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    event_id uuid,
    user_id uuid,
    status character varying DEFAULT 'registered'::character varying,
    registration_data jsonb,
    registered_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying NOT NULL,
    description text NOT NULL,
    club_id character varying,
    created_by uuid,
    event_date date NOT NULL,
    event_time time without time zone NOT NULL,
    location character varying NOT NULL,
    max_attendees integer,
    status character varying DEFAULT 'upcoming'::character varying,
    image_url text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    event_images jsonb DEFAULT '[]'::jsonb,
    banner_image_url text,
    gallery_images jsonb DEFAULT '[]'::jsonb
);


--
-- Name: likes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.likes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    post_id uuid,
    user_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: media_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.media_files (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    filename character varying NOT NULL,
    original_filename character varying NOT NULL,
    file_size integer NOT NULL,
    mime_type character varying NOT NULL,
    file_url text NOT NULL,
    thumbnail_url text,
    alt_text text,
    description text,
    uploaded_by uuid,
    upload_context character varying,
    upload_reference_id uuid,
    is_public boolean DEFAULT true,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    chat_room_id uuid NOT NULL,
    user_id uuid NOT NULL,
    content text NOT NULL,
    attachment_url text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id integer DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id uuid NOT NULL,
    type character varying NOT NULL,
    title text,
    message text NOT NULL,
    link text,
    read boolean DEFAULT false,
    delivery_method character varying DEFAULT 'in-app'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    sent_by character varying,
    club_id character varying,
    email_sent boolean DEFAULT false,
    email_sent_at timestamp without time zone,
    related_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb
);


--
-- Name: posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.posts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying NOT NULL,
    content text NOT NULL,
    author_id uuid,
    club_id character varying,
    category character varying DEFAULT 'blog'::character varying,
    post_type character varying DEFAULT 'blog'::character varying,
    tags text[] DEFAULT '{}'::text[],
    excerpt text,
    reading_time_minutes integer DEFAULT 0,
    featured_image_url text,
    post_images jsonb DEFAULT '[]'::jsonb,
    content_blocks jsonb DEFAULT '[]'::jsonb,
    meta_description text,
    slug character varying,
    status character varying DEFAULT 'draft'::character varying,
    is_featured boolean DEFAULT false,
    is_pinned boolean DEFAULT false,
    view_count integer DEFAULT 0,
    likes_count integer DEFAULT 0,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    edited_by uuid,
    search_vector tsvector DEFAULT to_tsvector('english'::regconfig, ''::text)
);


--
-- Name: proctoring_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.proctoring_sessions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    assignment_id uuid NOT NULL,
    user_id uuid NOT NULL,
    session_start timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    session_end timestamp with time zone,
    camera_enabled boolean DEFAULT false,
    microphone_enabled boolean DEFAULT false,
    face_verified boolean DEFAULT false,
    violations jsonb DEFAULT '[]'::jsonb,
    screenshots jsonb DEFAULT '[]'::jsonb,
    system_info jsonb DEFAULT '{}'::jsonb,
    session_data jsonb DEFAULT '{}'::jsonb
);


--
-- Name: project_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_invitations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    project_id uuid NOT NULL,
    inviter_id uuid NOT NULL,
    email character varying NOT NULL,
    role character varying DEFAULT 'member'::character varying,
    invitation_token character varying NOT NULL,
    project_password character varying,
    status character varying DEFAULT 'pending'::character varying,
    message text,
    expires_at timestamp with time zone DEFAULT (CURRENT_TIMESTAMP + '7 days'::interval) NOT NULL,
    sent_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: project_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_members (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role character varying DEFAULT 'member'::character varying,
    status character varying DEFAULT 'active'::character varying,
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    invited_by uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying NOT NULL,
    description text,
    club_id character varying NOT NULL,
    created_by uuid NOT NULL,
    project_key character varying NOT NULL,
    project_type character varying DEFAULT 'development'::character varying,
    priority character varying DEFAULT 'medium'::character varying,
    status character varying DEFAULT 'planning'::character varying,
    start_date date,
    target_end_date date,
    actual_end_date date,
    access_password character varying,
    is_public boolean DEFAULT false,
    progress_percentage numeric DEFAULT 0,
    total_tasks integer DEFAULT 0,
    completed_tasks integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: query_cache; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.query_cache (
    cache_key text NOT NULL,
    cache_value jsonb NOT NULL,
    last_updated timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL
);


--
-- Name: question_media; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.question_media (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    question_id uuid NOT NULL,
    media_file_id uuid NOT NULL,
    media_type character varying NOT NULL,
    display_order integer DEFAULT 0,
    is_primary boolean DEFAULT false,
    caption text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: question_options; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.question_options (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    question_id uuid NOT NULL,
    option_text text NOT NULL,
    is_correct boolean DEFAULT false,
    ordering integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: question_responses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.question_responses (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    submission_id uuid NOT NULL,
    question_id uuid NOT NULL,
    selected_options uuid[],
    code_answer text,
    essay_answer text,
    is_correct boolean,
    score integer DEFAULT 0,
    time_spent integer,
    feedback text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    selected_language character varying,
    last_auto_save timestamp with time zone,
    attempt_history jsonb DEFAULT '[]'::jsonb
);


--
-- Name: security_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.security_events (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    event_type character varying NOT NULL,
    ip_address character varying,
    device_info jsonb DEFAULT '{}'::jsonb,
    event_data jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    token character varying NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    last_active_at timestamp with time zone DEFAULT now(),
    user_agent text,
    ip_address character varying,
    device_info jsonb DEFAULT '{}'::jsonb,
    is_trusted boolean DEFAULT false,
    requires_2fa boolean DEFAULT true,
    has_completed_2fa boolean DEFAULT false
);


--
-- Name: system_statistics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_statistics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_statistics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_statistics (
    id integer DEFAULT nextval('public.system_statistics_id_seq'::regclass) NOT NULL,
    active_users_count integer DEFAULT 0,
    total_users_count integer DEFAULT 0,
    total_clubs_count integer DEFAULT 0,
    total_events_count integer DEFAULT 0,
    total_assignments_count integer DEFAULT 0,
    total_comments_count integer DEFAULT 0,
    daily_active_users integer DEFAULT 0,
    weekly_active_users integer DEFAULT 0,
    monthly_active_users integer DEFAULT 0,
    "timestamp" timestamp with time zone DEFAULT now()
);


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    project_id uuid NOT NULL,
    title character varying NOT NULL,
    description text,
    task_key character varying NOT NULL,
    task_type character varying DEFAULT 'task'::character varying,
    priority character varying DEFAULT 'medium'::character varying,
    status character varying DEFAULT 'todo'::character varying,
    assignee_id uuid,
    reporter_id uuid NOT NULL,
    parent_task_id uuid,
    story_points integer,
    time_spent_hours numeric DEFAULT 0,
    due_date timestamp with time zone,
    completed_date timestamp with time zone,
    is_completed boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: trusted_devices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trusted_devices (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    device_identifier character varying NOT NULL,
    device_name character varying NOT NULL,
    device_type character varying,
    browser character varying,
    os character varying,
    ip_address character varying,
    last_used timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone DEFAULT (CURRENT_TIMESTAMP + '30 days'::interval),
    trust_level character varying DEFAULT 'login_only'::character varying
);


--
-- Name: user_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_activities (
    id integer DEFAULT nextval('public.user_activities_id_seq'::regclass) NOT NULL,
    user_id uuid,
    action character varying NOT NULL,
    target_type character varying NOT NULL,
    target_id text,
    target_name text,
    details jsonb,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: user_badges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_badges (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    badge_name character varying NOT NULL,
    badge_description text,
    badge_icon character varying,
    earned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying NOT NULL,
    password_hash character varying NOT NULL,
    name character varying NOT NULL,
    username character varying,
    avatar text,
    role character varying DEFAULT 'student'::character varying NOT NULL,
    club_id character varying,
    bio text,
    social_links jsonb DEFAULT '{}'::jsonb,
    preferences jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    profile_image_url text,
    profile_images jsonb DEFAULT '[]'::jsonb,
    verification_photo_url text,
    phone_number character varying,
    date_of_birth date,
    address text,
    emergency_contact jsonb DEFAULT '{}'::jsonb,
    phone character varying,
    location character varying,
    website character varying,
    github character varying,
    linkedin character varying,
    twitter character varying,
    email_verified boolean DEFAULT false,
    email_verification_token character varying,
    email_verification_token_expires_at timestamp without time zone,
    password_reset_token character varying,
    password_reset_token_expires_at timestamp without time zone,
    oauth_provider character varying,
    oauth_id character varying,
    oauth_data jsonb,
    has_password boolean DEFAULT true,
    totp_secret character varying,
    totp_temp_secret character varying,
    totp_temp_secret_created_at timestamp without time zone,
    totp_enabled boolean DEFAULT false,
    totp_enabled_at timestamp without time zone,
    totp_recovery_codes jsonb,
    notification_preferences jsonb DEFAULT '{"email": {"events": true, "results": true, "assignments": true, "discussions": true}}'::jsonb,
    email_otp_enabled boolean DEFAULT false,
    email_otp_verified boolean DEFAULT false,
    email_otp_secret character varying,
    email_otp_backup_codes jsonb DEFAULT '[]'::jsonb,
    email_otp_last_used timestamp with time zone,
    email_otp_created_at timestamp with time zone,
    email_otp character(6),
    email_otp_expires_at timestamp with time zone,
    last_activity timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: ai_assignment_generations ai_assignment_generations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_assignment_generations
    ADD CONSTRAINT ai_assignment_generations_pkey PRIMARY KEY (id);


--
-- Name: announcements announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.announcements
    ADD CONSTRAINT announcements_pkey PRIMARY KEY (id);


--
-- Name: assignment_attempts assignment_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignment_attempts
    ADD CONSTRAINT assignment_attempts_pkey PRIMARY KEY (id);


--
-- Name: assignment_audit_log assignment_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignment_audit_log
    ADD CONSTRAINT assignment_audit_log_pkey PRIMARY KEY (id);


--
-- Name: assignment_questions assignment_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignment_questions
    ADD CONSTRAINT assignment_questions_pkey PRIMARY KEY (id);


--
-- Name: assignment_submissions assignment_submissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignment_submissions
    ADD CONSTRAINT assignment_submissions_pkey PRIMARY KEY (id);


--
-- Name: assignment_templates assignment_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignment_templates
    ADD CONSTRAINT assignment_templates_pkey PRIMARY KEY (id);


--
-- Name: assignment_violations assignment_violations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignment_violations
    ADD CONSTRAINT assignment_violations_pkey PRIMARY KEY (id);


--
-- Name: assignments assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assignments
    ADD CONSTRAINT assignments_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: chat_attachments chat_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_attachments
    ADD CONSTRAINT chat_attachments_pkey PRIMARY KEY (id);


--
-- Name: chat_invitations chat_invitations_invitation_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_invitations
    ADD CONSTRAINT chat_invitations_invitation_token_key UNIQUE (invitation_token);


--
-- Name: chat_invitations chat_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_invitations
    ADD CONSTRAINT chat_invitations_pkey PRIMARY KEY (id);


--
-- Name: chat_messages chat_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_pkey PRIMARY KEY (id);


--
-- Name: chat_room_members chat_room_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_room_members
    ADD CONSTRAINT chat_room_members_pkey PRIMARY KEY (id);


--
-- Name: chat_rooms chat_rooms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_rooms
    ADD CONSTRAINT chat_rooms_pkey PRIMARY KEY (id);


--
-- Name: club_members club_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.club_members
    ADD CONSTRAINT club_members_pkey PRIMARY KEY (id);


--
-- Name: club_statistics club_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.club_statistics
    ADD CONSTRAINT club_statistics_pkey PRIMARY KEY (id);


--
-- Name: clubs clubs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_pkey PRIMARY KEY (id);


--
-- Name: code_results code_results_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.code_results
    ADD CONSTRAINT code_results_pkey PRIMARY KEY (id);


--
-- Name: coding_submissions coding_submissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coding_submissions
    ADD CONSTRAINT coding_submissions_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: committee_members committee_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.committee_members
    ADD CONSTRAINT committee_members_pkey PRIMARY KEY (id);


--
-- Name: committee_roles committee_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.committee_roles
    ADD CONSTRAINT committee_roles_pkey PRIMARY KEY (id);


--
-- Name: committees committees_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.committees
    ADD CONSTRAINT committees_name_key UNIQUE (name);


--
-- Name: committees committees_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.committees
    ADD CONSTRAINT committees_pkey PRIMARY KEY (id);


--
-- Name: discussion_replies discussion_replies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussion_replies
    ADD CONSTRAINT discussion_replies_pkey PRIMARY KEY (id);


--
-- Name: discussions discussions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussions
    ADD CONSTRAINT discussions_pkey PRIMARY KEY (id);


--
-- Name: email_logs email_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_logs
    ADD CONSTRAINT email_logs_pkey PRIMARY KEY (id);


--
-- Name: event_attendees event_attendees_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_attendees
    ADD CONSTRAINT event_attendees_pkey PRIMARY KEY (id);


--
-- Name: event_registrations event_registrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_registrations
    ADD CONSTRAINT event_registrations_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: likes likes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT likes_pkey PRIMARY KEY (id);


--
-- Name: media_files media_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_files
    ADD CONSTRAINT media_files_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: posts posts_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_slug_key UNIQUE (slug);


--
-- Name: proctoring_sessions proctoring_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proctoring_sessions
    ADD CONSTRAINT proctoring_sessions_pkey PRIMARY KEY (id);


--
-- Name: project_invitations project_invitations_invitation_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_invitations
    ADD CONSTRAINT project_invitations_invitation_token_key UNIQUE (invitation_token);


--
-- Name: project_invitations project_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_invitations
    ADD CONSTRAINT project_invitations_pkey PRIMARY KEY (id);


--
-- Name: project_members project_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: query_cache query_cache_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.query_cache
    ADD CONSTRAINT query_cache_pkey PRIMARY KEY (cache_key);


--
-- Name: question_media question_media_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.question_media
    ADD CONSTRAINT question_media_pkey PRIMARY KEY (id);


--
-- Name: question_options question_options_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.question_options
    ADD CONSTRAINT question_options_pkey PRIMARY KEY (id);


--
-- Name: question_responses question_responses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.question_responses
    ADD CONSTRAINT question_responses_pkey PRIMARY KEY (id);


--
-- Name: security_events security_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.security_events
    ADD CONSTRAINT security_events_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_token_key UNIQUE (token);


--
-- Name: system_statistics system_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_statistics
    ADD CONSTRAINT system_statistics_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: trusted_devices trusted_devices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trusted_devices
    ADD CONSTRAINT trusted_devices_pkey PRIMARY KEY (id);


--
-- Name: likes unique_post_user_like; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT unique_post_user_like UNIQUE (post_id, user_id);


--
-- Name: projects unique_project_key_club; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT unique_project_key_club UNIQUE (club_id, project_key);


--
-- Name: project_members unique_project_member; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT unique_project_member UNIQUE (project_id, user_id);


--
-- Name: tasks unique_task_key_project; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT unique_task_key_project UNIQUE (project_id, task_key);


--
-- Name: user_activities user_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_activities
    ADD CONSTRAINT user_activities_pkey PRIMARY KEY (id);


--
-- Name: user_badges user_badges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_badges
    ADD CONSTRAINT user_badges_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_ai_generations_assignment_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_generations_assignment_id ON public.ai_assignment_generations USING btree (generated_assignment_id);


--
-- Name: idx_ai_generations_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_generations_created_at ON public.ai_assignment_generations USING btree (created_at);


--
-- Name: idx_ai_generations_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_generations_status ON public.ai_assignment_generations USING btree (generation_status);


--
-- Name: idx_ai_generations_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_generations_template_id ON public.ai_assignment_generations USING btree (template_id);


--
-- Name: idx_assignment_attempts_assignment_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_assignment_id ON public.assignment_attempts USING btree (assignment_id);


--
-- Name: idx_assignment_attempts_assignment_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_assignment_user ON public.assignment_attempts USING btree (assignment_id, user_id);


--
-- Name: idx_assignment_attempts_auto_save; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_auto_save ON public.assignment_attempts USING btree (assignment_id, user_id, last_auto_save);


--
-- Name: idx_assignment_attempts_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_status ON public.assignment_attempts USING btree (status);


--
-- Name: idx_assignment_attempts_submitted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_submitted_at ON public.assignment_attempts USING btree (submitted_at);


--
-- Name: idx_assignment_attempts_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_user_id ON public.assignment_attempts USING btree (user_id);


--
-- Name: idx_assignment_attempts_user_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_user_status ON public.assignment_attempts USING btree (user_id, status, submitted_at DESC);


--
-- Name: idx_assignment_attempts_violations; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_attempts_violations ON public.assignment_attempts USING btree (assignment_id, window_violations);


--
-- Name: idx_assignment_audit_log_assignment; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_audit_log_assignment ON public.assignment_audit_log USING btree (assignment_id);


--
-- Name: idx_assignment_questions_assignment_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_questions_assignment_id ON public.assignment_questions USING btree (assignment_id);


--
-- Name: idx_assignment_questions_correct_answer_jsonb; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_questions_correct_answer_jsonb ON public.assignment_questions USING gin (correct_answer);


--
-- Name: idx_assignment_questions_language_settings; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_questions_language_settings ON public.assignment_questions USING btree (code_language, allow_any_language);


--
-- Name: idx_assignment_submissions_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_assignment_submissions_unique ON public.assignment_submissions USING btree (assignment_id, user_id);


--
-- Name: idx_assignment_submissions_user_submitted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_submissions_user_submitted ON public.assignment_submissions USING btree (user_id, submitted_at DESC);


--
-- Name: idx_assignment_templates_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_templates_category ON public.assignment_templates USING btree (category);


--
-- Name: idx_assignment_templates_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_templates_created_by ON public.assignment_templates USING btree (created_by);


--
-- Name: idx_assignment_templates_difficulty; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_templates_difficulty ON public.assignment_templates USING btree (difficulty_level);


--
-- Name: idx_assignment_templates_subject; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_templates_subject ON public.assignment_templates USING btree (subject);


--
-- Name: idx_assignment_violations_submission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignment_violations_submission_id ON public.assignment_violations USING btree (submission_id);


--
-- Name: idx_assignments_club_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignments_club_id ON public.assignments USING btree (club_id);


--
-- Name: idx_assignments_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignments_created_by ON public.assignments USING btree (created_by);


--
-- Name: idx_assignments_due_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignments_due_date ON public.assignments USING btree (due_date);


--
-- Name: idx_assignments_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignments_status ON public.assignments USING btree (status);


--
-- Name: idx_assignments_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignments_type ON public.assignments USING btree (assignment_type);


--
-- Name: idx_attempts_assignment_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attempts_assignment_status ON public.assignment_attempts USING btree (assignment_id, status);


--
-- Name: idx_attempts_user_assignment; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attempts_user_assignment ON public.assignment_attempts USING btree (user_id, assignment_id, attempt_number);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_resource_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_resource_type ON public.audit_logs USING btree (resource_type);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_chat_messages_can_edit_until; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_can_edit_until ON public.chat_messages USING btree (can_edit_until);


--
-- Name: idx_chat_messages_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_created_at ON public.chat_messages USING btree (created_at);


--
-- Name: idx_chat_messages_edited_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_edited_at ON public.chat_messages USING btree (edited_at);


--
-- Name: idx_chat_messages_room_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_room_id ON public.chat_messages USING btree (room_id);


--
-- Name: idx_chat_messages_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_user_id ON public.chat_messages USING btree (user_id);


--
-- Name: idx_chat_room_members_chat_room_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_room_members_chat_room_id ON public.chat_room_members USING btree (chat_room_id);


--
-- Name: idx_chat_room_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_room_members_user_id ON public.chat_room_members USING btree (user_id);


--
-- Name: idx_chat_rooms_profile_picture; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_rooms_profile_picture ON public.chat_rooms USING btree (profile_picture_url);


--
-- Name: idx_club_members_club_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_club_members_club_id ON public.club_members USING btree (club_id);


--
-- Name: idx_club_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_club_members_user_id ON public.club_members USING btree (user_id);


--
-- Name: idx_clubs_coordinator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_clubs_coordinator_id ON public.clubs USING btree (coordinator_id);


--
-- Name: idx_clubs_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_clubs_type ON public.clubs USING btree (type);


--
-- Name: idx_coding_submissions_question_response_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_coding_submissions_question_response_id ON public.coding_submissions USING btree (question_response_id);


--
-- Name: idx_comments_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_comments_author_id ON public.comments USING btree (author_id);


--
-- Name: idx_comments_post_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_comments_post_id ON public.comments USING btree (post_id);


--
-- Name: idx_committee_members_committee_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_committee_members_committee_id ON public.committee_members USING btree (committee_id);


--
-- Name: idx_committee_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_committee_members_user_id ON public.committee_members USING btree (user_id);


--
-- Name: idx_committee_roles_committee_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_committee_roles_committee_id ON public.committee_roles USING btree (committee_id);


--
-- Name: idx_email_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_logs_created_at ON public.email_logs USING btree (created_at);


--
-- Name: idx_email_logs_recipient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_logs_recipient ON public.email_logs USING btree (recipient);


--
-- Name: idx_email_logs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_logs_status ON public.email_logs USING btree (status);


--
-- Name: idx_event_attendees_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_attendees_event_id ON public.event_attendees USING btree (event_id);


--
-- Name: idx_event_attendees_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_attendees_user_id ON public.event_attendees USING btree (user_id);


--
-- Name: idx_events_club_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_club_id ON public.events USING btree (club_id);


--
-- Name: idx_events_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_created_by ON public.events USING btree (created_by);


--
-- Name: idx_events_event_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_event_date ON public.events USING btree (event_date);


--
-- Name: idx_likes_post_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_likes_post_id ON public.likes USING btree (post_id);


--
-- Name: idx_likes_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_likes_user_id ON public.likes USING btree (user_id);


--
-- Name: idx_media_files_upload_context; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_media_files_upload_context ON public.media_files USING btree (upload_context);


--
-- Name: idx_media_files_uploaded_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_media_files_uploaded_by ON public.media_files USING btree (uploaded_by);


--
-- Name: idx_notifications_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_created_at ON public.notifications USING btree (created_at);


--
-- Name: idx_notifications_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_read ON public.notifications USING btree (read);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_id ON public.notifications USING btree (user_id);


--
-- Name: idx_posts_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_posts_author_id ON public.posts USING btree (author_id);


--
-- Name: idx_posts_club_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_posts_club_id ON public.posts USING btree (club_id);


--
-- Name: idx_posts_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_posts_created_at ON public.posts USING btree (created_at);


--
-- Name: idx_posts_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_posts_search_vector ON public.posts USING gin (search_vector);


--
-- Name: idx_posts_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_posts_slug ON public.posts USING btree (slug);


--
-- Name: idx_posts_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_posts_status ON public.posts USING btree (status);


--
-- Name: idx_project_members_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_members_project_id ON public.project_members USING btree (project_id);


--
-- Name: idx_project_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_members_user_id ON public.project_members USING btree (user_id);


--
-- Name: idx_projects_club_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_club_id ON public.projects USING btree (club_id);


--
-- Name: idx_projects_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_created_by ON public.projects USING btree (created_by);


--
-- Name: idx_projects_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_status ON public.projects USING btree (status);


--
-- Name: idx_question_responses_question_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_question_responses_question_id ON public.question_responses USING btree (question_id);


--
-- Name: idx_question_responses_submission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_question_responses_submission_id ON public.question_responses USING btree (submission_id);


--
-- Name: idx_questions_assignment_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_questions_assignment_order ON public.assignment_questions USING btree (assignment_id, question_order);


--
-- Name: idx_questions_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_questions_type ON public.assignment_questions USING btree (question_type);


--
-- Name: idx_security_events_event_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_security_events_event_type ON public.security_events USING btree (event_type);


--
-- Name: idx_security_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_security_events_user_id ON public.security_events USING btree (user_id);


--
-- Name: idx_sessions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_expires_at ON public.sessions USING btree (expires_at);


--
-- Name: idx_sessions_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_token ON public.sessions USING btree (token);


--
-- Name: idx_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user_id ON public.sessions USING btree (user_id);


--
-- Name: idx_submissions_assignment_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_assignment_status ON public.assignment_submissions USING btree (assignment_id, status);


--
-- Name: idx_submissions_status_submitted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_status_submitted ON public.assignment_submissions USING btree (status, submitted_at DESC);


--
-- Name: idx_submissions_user_submitted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_user_submitted ON public.assignment_submissions USING btree (user_id, submitted_at DESC);


--
-- Name: idx_tasks_assignee_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_assignee_id ON public.tasks USING btree (assignee_id);


--
-- Name: idx_tasks_due_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_due_date ON public.tasks USING btree (due_date);


--
-- Name: idx_tasks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_project_id ON public.tasks USING btree (project_id);


--
-- Name: idx_tasks_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_status ON public.tasks USING btree (status);


--
-- Name: idx_trusted_devices_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trusted_devices_user_id ON public.trusted_devices USING btree (user_id);


--
-- Name: idx_user_activities_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_activities_created_at ON public.user_activities USING btree (created_at);


--
-- Name: idx_user_activities_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_activities_user_id ON public.user_activities USING btree (user_id);


--
-- Name: idx_users_club_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_club_id ON public.users USING btree (club_id);


--
-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_created_at ON public.users USING btree (created_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_last_activity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_last_activity ON public.users USING btree (last_activity);


--
-- Name: idx_users_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_role ON public.users USING btree (role);


--
-- Name: posts posts_search_vector_update; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER posts_search_vector_update BEFORE INSERT OR UPDATE ON public.posts FOR EACH ROW EXECUTE FUNCTION public.update_post_search_vector();


--
-- Name: tasks tasks_generate_key; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER tasks_generate_key BEFORE INSERT ON public.tasks FOR EACH ROW WHEN (((new.task_key IS NULL) OR ((new.task_key)::text = ''::text))) EXECUTE FUNCTION public.generate_task_key();


--
-- Name: tasks tasks_update_project_progress; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER tasks_update_project_progress AFTER INSERT OR DELETE OR UPDATE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.update_project_progress();


--
-- Name: assignments update_assignments_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_assignments_updated_at BEFORE UPDATE ON public.assignments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: clubs update_clubs_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_clubs_updated_at BEFORE UPDATE ON public.clubs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: comments update_comments_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: committee_members update_committee_members_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_committee_members_updated_at BEFORE UPDATE ON public.committee_members FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: committee_roles update_committee_roles_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_committee_roles_updated_at BEFORE UPDATE ON public.committee_roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: committees update_committees_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_committees_updated_at BEFORE UPDATE ON public.committees FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: events update_events_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON public.events FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: posts update_posts_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON public.posts FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: projects update_projects_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: tasks update_tasks_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: chat_messages fk_chat_messages_edited_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT fk_chat_messages_edited_by FOREIGN KEY (edited_by) REFERENCES public.users(id);


--
-- Name: chat_rooms fk_chat_rooms_edited_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_rooms
    ADD CONSTRAINT fk_chat_rooms_edited_by FOREIGN KEY (edited_by) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

