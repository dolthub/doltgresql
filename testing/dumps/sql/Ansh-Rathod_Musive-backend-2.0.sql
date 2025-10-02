-- Downloaded from: https://github.com/Ansh-Rathod/Musive-backend-2.0/blob/eb320d80d2fa07283bb4ab9351581f4c8757bcad/schema.sql
create table Users(
    id serial primary key,
    username varchar(28) not null unique,
    passhash varchar not null
);
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table public."Artists"(
  id integer unique not null, 
  username text not null unique,
  display_name text not null,
  avatar jsonb,
  gender varchar,
  PRIMARY KEY(id)
);

create table public."Tracks"(
  id integer unique not null,
  user_id integer not null,
  tags text[] not null DEFAULT '{}',
  moods text[] not null DEFAULT '{}',
  genres text[] not null DEFAULT '{}',
  movements text[] not null DEFAULT '{}',
  keywords text not null,
  duration float not null,
  track_name text not null,
  download_url text not null,
  src text not null,
  cover_image jsonb,
  PRIMARY KEY(id)
);

alter table songs
add constraint user_id_fk FOREIGN KEY (user_id) REFERENCES artists(id)
match full on update CASCADE on delete CASCADE;

create table public."Liked"(
    id serial primary key,
    track_id integer not null,
    username varchar(28) not null
);

alter table public."Liked" add constraint track_id FOREIGN KEY(track_id)
REFERENCES public."Tracks"(id) match full on update CASCADE on delete cascade;

alter table public."Liked" add constraint user_id FOREIGN KEY(username)
REFERENCES public."Users"(username) match full on update CASCADE on delete cascade;

create table public."Collections"(
  id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
  name text not null,
  username varchar(28) not null,
  total_tracks integer DEFAULT 0
)

alter table public."Collections" add constraint user_id FOREIGN KEY(username)
REFERENCES public."Users"(username) match full on update CASCADE on delete cascade;

create table public."CollectionItems"(
  collection_id uuid not null,
  track_id integer not null
);


alter table public."CollectionItems" add constraint collection_id_fk FOREIGN KEY(collection_id)
REFERENCES public."Collections"(id) match full on update CASCADE on delete cascade;

alter table public."CollectionItems" add constraint track_id_fk FOREIGN KEY(track_id)
REFERENCES public."Tracks"(id) match full on update CASCADE on delete cascade;


CREATE OR REPLACE FUNCTION update_collections()
  RETURNS trigger AS $$
  DECLARE
    BEGIN
    IF TG_OP = 'INSERT' THEN
      EXECUTE 'update public."Collections" set total_tracks=total_tracks+1 where id = $1;' 
      USING NEW.collection_id;
    END IF;

    IF TG_OP = 'DELETE' THEN 
      EXECUTE 'update public."Collections" set total_tracks=total_tracks-1 where id = $1;' 
      USING OLD.collection_id;
    END IF;
    
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

  
CREATE TRIGGER update_collection
AFTER INSERT OR DELETE ON public."CollectionItems"
FOR EACH ROW EXECUTE PROCEDURE update_collections();

-- pg_dump -U postgres -h containers-us-west-63.railway.app -p 7771 railway >> sqlfile.sql