// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPlpgsql(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_plpgsql)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_plpgsql,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `--     of the wall connectors is wired to one of several patch-
create table Room (
    roomno	char(8),
    comment	text
);`,
			},
			{
				Statement: `create unique index Room_rno on Room using btree (roomno bpchar_ops);`,
			},
			{
				Statement: `create table WSlot (
    slotname	char(20),
    roomno	char(8),
    slotlink	char(20),
    backlink	char(20)
);`,
			},
			{
				Statement: `create unique index WSlot_name on WSlot using btree (slotname bpchar_ops);`,
			},
			{
				Statement: `create table PField (
    name	text,
    comment	text
);`,
			},
			{
				Statement: `create unique index PField_name on PField using btree (name text_ops);`,
			},
			{
				Statement: `create table PSlot (
    slotname	char(20),
    pfname	text,
    slotlink	char(20),
    backlink	char(20)
);`,
			},
			{
				Statement: `create unique index PSlot_name on PSlot using btree (slotname bpchar_ops);`,
			},
			{
				Statement: `create table PLine (
    slotname	char(20),
    phonenumber	char(20),
    comment	text,
    backlink	char(20)
);`,
			},
			{
				Statement: `create unique index PLine_name on PLine using btree (slotname bpchar_ops);`,
			},
			{
				Statement: `create table Hub (
    name	char(14),
    comment	text,
    nslots	integer
);`,
			},
			{
				Statement: `create unique index Hub_name on Hub using btree (name bpchar_ops);`,
			},
			{
				Statement: `create table HSlot (
    slotname	char(20),
    hubname	char(14),
    slotno	integer,
    slotlink	char(20)
);`,
			},
			{
				Statement: `create unique index HSlot_name on HSlot using btree (slotname bpchar_ops);`,
			},
			{
				Statement: `create index HSlot_hubname on HSlot using btree (hubname bpchar_ops);`,
			},
			{
				Statement: `create table System (
    name	text,
    comment	text
);`,
			},
			{
				Statement: `create unique index System_name on System using btree (name text_ops);`,
			},
			{
				Statement: `create table IFace (
    slotname	char(20),
    sysname	text,
    ifname	text,
    slotlink	char(20)
);`,
			},
			{
				Statement: `create unique index IFace_name on IFace using btree (slotname bpchar_ops);`,
			},
			{
				Statement: `create table PHone (
    slotname	char(20),
    comment	text,
    slotlink	char(20)
);`,
			},
			{
				Statement: `create unique index PHone_name on PHone using btree (slotname bpchar_ops);`,
			},
			{
				Statement: `create function tg_room_au() returns trigger as '
begin
    if new.roomno != old.roomno then
        update WSlot set roomno = new.roomno where roomno = old.roomno;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_room_au after update
    on Room for each row execute procedure tg_room_au();`,
			},
			{
				Statement: `create function tg_room_ad() returns trigger as '
begin
    delete from WSlot where roomno = old.roomno;`,
			},
			{
				Statement: `    return old;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_room_ad after delete
    on Room for each row execute procedure tg_room_ad();`,
			},
			{
				Statement: `create function tg_wslot_biu() returns trigger as $$
begin
    if count(*) = 0 from Room where roomno = new.roomno then
        raise exception 'Room % does not exist', new.roomno;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create trigger tg_wslot_biu before insert or update
    on WSlot for each row execute procedure tg_wslot_biu();`,
			},
			{
				Statement: `create function tg_pfield_au() returns trigger as '
begin
    if new.name != old.name then
        update PSlot set pfname = new.name where pfname = old.name;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_pfield_au after update
    on PField for each row execute procedure tg_pfield_au();`,
			},
			{
				Statement: `create function tg_pfield_ad() returns trigger as '
begin
    delete from PSlot where pfname = old.name;`,
			},
			{
				Statement: `    return old;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_pfield_ad after delete
    on PField for each row execute procedure tg_pfield_ad();`,
			},
			{
				Statement: `create function tg_pslot_biu() returns trigger as $proc$
declare
    pfrec	record;`,
			},
			{
				Statement: `    ps          alias for new;`,
			},
			{
				Statement: `begin
    select into pfrec * from PField where name = ps.pfname;`,
			},
			{
				Statement: `    if not found then
        raise exception $$Patchfield "%" does not exist$$, ps.pfname;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return ps;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$proc$ language plpgsql;`,
			},
			{
				Statement: `create trigger tg_pslot_biu before insert or update
    on PSlot for each row execute procedure tg_pslot_biu();`,
			},
			{
				Statement: `create function tg_system_au() returns trigger as '
begin
    if new.name != old.name then
        update IFace set sysname = new.name where sysname = old.name;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_system_au after update
    on System for each row execute procedure tg_system_au();`,
			},
			{
				Statement: `create function tg_iface_biu() returns trigger as $$
declare
    sname	text;`,
			},
			{
				Statement: `    sysrec	record;`,
			},
			{
				Statement: `begin
    select into sysrec * from system where name = new.sysname;`,
			},
			{
				Statement: `    if not found then
        raise exception $q$system "%" does not exist$q$, new.sysname;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    sname := 'IF.' || new.sysname;`,
			},
			{
				Statement: `    sname := sname || '.';`,
			},
			{
				Statement: `    sname := sname || new.ifname;`,
			},
			{
				Statement: `    if length(sname) > 20 then
        raise exception 'IFace slotname "%" too long (20 char max)', sname;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    new.slotname := sname;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create trigger tg_iface_biu before insert or update
    on IFace for each row execute procedure tg_iface_biu();`,
			},
			{
				Statement: `create function tg_hub_a() returns trigger as '
declare
    hname	text;`,
			},
			{
				Statement: `    dummy	integer;`,
			},
			{
				Statement: `begin
    if tg_op = ''INSERT'' then
	dummy := tg_hub_adjustslots(new.name, 0, new.nslots);`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''UPDATE'' then
	if new.name != old.name then
	    update HSlot set hubname = new.name where hubname = old.name;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	dummy := tg_hub_adjustslots(new.name, old.nslots, new.nslots);`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''DELETE'' then
	dummy := tg_hub_adjustslots(old.name, old.nslots, 0);`,
			},
			{
				Statement: `	return old;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_hub_a after insert or update or delete
    on Hub for each row execute procedure tg_hub_a();`,
			},
			{
				Statement: `create function tg_hub_adjustslots(hname bpchar,
                                   oldnslots integer,
                                   newnslots integer)
returns integer as '
begin
    if newnslots = oldnslots then
        return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if newnslots < oldnslots then
        delete from HSlot where hubname = hname and slotno > newnslots;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    for i in oldnslots + 1 .. newnslots loop
        insert into HSlot (slotname, hubname, slotno, slotlink)
		values (''HS.dummy'', hname, i, '''');`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    return 0;`,
			},
			{
				Statement: `end
' language plpgsql;`,
			},
			{
				Statement:   `COMMENT ON FUNCTION tg_hub_adjustslots_wrong(bpchar, integer, integer) IS 'function with args';`,
				ErrorString: `function tg_hub_adjustslots_wrong(character, integer, integer) does not exist`,
			},
			{
				Statement: `COMMENT ON FUNCTION tg_hub_adjustslots(bpchar, integer, integer) IS 'function with args';`,
			},
			{
				Statement: `COMMENT ON FUNCTION tg_hub_adjustslots(bpchar, integer, integer) IS NULL;`,
			},
			{
				Statement: `create function tg_hslot_biu() returns trigger as '
declare
    sname	text;`,
			},
			{
				Statement: `    xname	HSlot.slotname%TYPE;`,
			},
			{
				Statement: `    hubrec	record;`,
			},
			{
				Statement: `begin
    select into hubrec * from Hub where name = new.hubname;`,
			},
			{
				Statement: `    if not found then
        raise exception ''no manual manipulation of HSlot'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if new.slotno < 1 or new.slotno > hubrec.nslots then
        raise exception ''no manual manipulation of HSlot'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''UPDATE'' and new.hubname != old.hubname then
	if count(*) > 0 from Hub where name = old.hubname then
	    raise exception ''no manual manipulation of HSlot'';`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    sname := ''HS.'' || trim(new.hubname);`,
			},
			{
				Statement: `    sname := sname || ''.'';`,
			},
			{
				Statement: `    sname := sname || new.slotno::text;`,
			},
			{
				Statement: `    if length(sname) > 20 then
        raise exception ''HSlot slotname "%" too long (20 char max)'', sname;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    new.slotname := sname;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_hslot_biu before insert or update
    on HSlot for each row execute procedure tg_hslot_biu();`,
			},
			{
				Statement: `create function tg_hslot_bd() returns trigger as '
declare
    hubrec	record;`,
			},
			{
				Statement: `begin
    select into hubrec * from Hub where name = old.hubname;`,
			},
			{
				Statement: `    if not found then
        return old;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if old.slotno > hubrec.nslots then
        return old;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    raise exception ''no manual manipulation of HSlot'';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_hslot_bd before delete
    on HSlot for each row execute procedure tg_hslot_bd();`,
			},
			{
				Statement: `create function tg_chkslotname() returns trigger as '
begin
    if substr(new.slotname, 1, 2) != tg_argv[0] then
        raise exception ''slotname must begin with %'', tg_argv[0];`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_chkslotname before insert
    on PSlot for each row execute procedure tg_chkslotname('PS');`,
			},
			{
				Statement: `create trigger tg_chkslotname before insert
    on WSlot for each row execute procedure tg_chkslotname('WS');`,
			},
			{
				Statement: `create trigger tg_chkslotname before insert
    on PLine for each row execute procedure tg_chkslotname('PL');`,
			},
			{
				Statement: `create trigger tg_chkslotname before insert
    on IFace for each row execute procedure tg_chkslotname('IF');`,
			},
			{
				Statement: `create trigger tg_chkslotname before insert
    on PHone for each row execute procedure tg_chkslotname('PH');`,
			},
			{
				Statement: `create function tg_chkslotlink() returns trigger as '
begin
    if new.slotlink isnull then
        new.slotlink := '''';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_chkslotlink before insert or update
    on PSlot for each row execute procedure tg_chkslotlink();`,
			},
			{
				Statement: `create trigger tg_chkslotlink before insert or update
    on WSlot for each row execute procedure tg_chkslotlink();`,
			},
			{
				Statement: `create trigger tg_chkslotlink before insert or update
    on IFace for each row execute procedure tg_chkslotlink();`,
			},
			{
				Statement: `create trigger tg_chkslotlink before insert or update
    on HSlot for each row execute procedure tg_chkslotlink();`,
			},
			{
				Statement: `create trigger tg_chkslotlink before insert or update
    on PHone for each row execute procedure tg_chkslotlink();`,
			},
			{
				Statement: `create function tg_chkbacklink() returns trigger as '
begin
    if new.backlink isnull then
        new.backlink := '''';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_chkbacklink before insert or update
    on PSlot for each row execute procedure tg_chkbacklink();`,
			},
			{
				Statement: `create trigger tg_chkbacklink before insert or update
    on WSlot for each row execute procedure tg_chkbacklink();`,
			},
			{
				Statement: `create trigger tg_chkbacklink before insert or update
    on PLine for each row execute procedure tg_chkbacklink();`,
			},
			{
				Statement: `create function tg_pslot_bu() returns trigger as '
begin
    if new.slotname != old.slotname then
        delete from PSlot where slotname = old.slotname;`,
			},
			{
				Statement: `	insert into PSlot (
		    slotname,
		    pfname,
		    slotlink,
		    backlink
		) values (
		    new.slotname,
		    new.pfname,
		    new.slotlink,
		    new.backlink
		);`,
			},
			{
				Statement: `        return null;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_pslot_bu before update
    on PSlot for each row execute procedure tg_pslot_bu();`,
			},
			{
				Statement: `create function tg_wslot_bu() returns trigger as '
begin
    if new.slotname != old.slotname then
        delete from WSlot where slotname = old.slotname;`,
			},
			{
				Statement: `	insert into WSlot (
		    slotname,
		    roomno,
		    slotlink,
		    backlink
		) values (
		    new.slotname,
		    new.roomno,
		    new.slotlink,
		    new.backlink
		);`,
			},
			{
				Statement: `        return null;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_wslot_bu before update
    on WSlot for each row execute procedure tg_Wslot_bu();`,
			},
			{
				Statement: `create function tg_pline_bu() returns trigger as '
begin
    if new.slotname != old.slotname then
        delete from PLine where slotname = old.slotname;`,
			},
			{
				Statement: `	insert into PLine (
		    slotname,
		    phonenumber,
		    comment,
		    backlink
		) values (
		    new.slotname,
		    new.phonenumber,
		    new.comment,
		    new.backlink
		);`,
			},
			{
				Statement: `        return null;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_pline_bu before update
    on PLine for each row execute procedure tg_pline_bu();`,
			},
			{
				Statement: `create function tg_iface_bu() returns trigger as '
begin
    if new.slotname != old.slotname then
        delete from IFace where slotname = old.slotname;`,
			},
			{
				Statement: `	insert into IFace (
		    slotname,
		    sysname,
		    ifname,
		    slotlink
		) values (
		    new.slotname,
		    new.sysname,
		    new.ifname,
		    new.slotlink
		);`,
			},
			{
				Statement: `        return null;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_iface_bu before update
    on IFace for each row execute procedure tg_iface_bu();`,
			},
			{
				Statement: `create function tg_hslot_bu() returns trigger as '
begin
    if new.slotname != old.slotname or new.hubname != old.hubname then
        delete from HSlot where slotname = old.slotname;`,
			},
			{
				Statement: `	insert into HSlot (
		    slotname,
		    hubname,
		    slotno,
		    slotlink
		) values (
		    new.slotname,
		    new.hubname,
		    new.slotno,
		    new.slotlink
		);`,
			},
			{
				Statement: `        return null;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_hslot_bu before update
    on HSlot for each row execute procedure tg_hslot_bu();`,
			},
			{
				Statement: `create function tg_phone_bu() returns trigger as '
begin
    if new.slotname != old.slotname then
        delete from PHone where slotname = old.slotname;`,
			},
			{
				Statement: `	insert into PHone (
		    slotname,
		    comment,
		    slotlink
		) values (
		    new.slotname,
		    new.comment,
		    new.slotlink
		);`,
			},
			{
				Statement: `        return null;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_phone_bu before update
    on PHone for each row execute procedure tg_phone_bu();`,
			},
			{
				Statement: `create function tg_backlink_a() returns trigger as '
declare
    dummy	integer;`,
			},
			{
				Statement: `begin
    if tg_op = ''INSERT'' then
        if new.backlink != '''' then
	    dummy := tg_backlink_set(new.backlink, new.slotname);`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''UPDATE'' then
        if new.backlink != old.backlink then
	    if old.backlink != '''' then
	        dummy := tg_backlink_unset(old.backlink, old.slotname);`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	    if new.backlink != '''' then
	        dummy := tg_backlink_set(new.backlink, new.slotname);`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	else
	    if new.slotname != old.slotname and new.backlink != '''' then
	        dummy := tg_slotlink_set(new.backlink, new.slotname);`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''DELETE'' then
        if old.backlink != '''' then
	    dummy := tg_backlink_unset(old.backlink, old.slotname);`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return old;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_backlink_a after insert or update or delete
    on PSlot for each row execute procedure tg_backlink_a('PS');`,
			},
			{
				Statement: `create trigger tg_backlink_a after insert or update or delete
    on WSlot for each row execute procedure tg_backlink_a('WS');`,
			},
			{
				Statement: `create trigger tg_backlink_a after insert or update or delete
    on PLine for each row execute procedure tg_backlink_a('PL');`,
			},
			{
				Statement: `create function tg_backlink_set(myname bpchar, blname bpchar)
returns integer as '
declare
    mytype	char(2);`,
			},
			{
				Statement: `    link	char(4);`,
			},
			{
				Statement: `    rec		record;`,
			},
			{
				Statement: `begin
    mytype := substr(myname, 1, 2);`,
			},
			{
				Statement: `    link := mytype || substr(blname, 1, 2);`,
			},
			{
				Statement: `    if link = ''PLPL'' then
        raise exception
		''backlink between two phone lines does not make sense'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if link in (''PLWS'', ''WSPL'') then
        raise exception
		''direct link of phone line to wall slot not permitted'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''PS'' then
        select into rec * from PSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.backlink != blname then
	    update PSlot set backlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''WS'' then
        select into rec * from WSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.backlink != blname then
	    update WSlot set backlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''PL'' then
        select into rec * from PLine where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.backlink != blname then
	    update PLine set backlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    raise exception ''illegal backlink beginning with %'', mytype;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create function tg_backlink_unset(bpchar, bpchar)
returns integer as '
declare
    myname	alias for $1;`,
			},
			{
				Statement: `    blname	alias for $2;`,
			},
			{
				Statement: `    mytype	char(2);`,
			},
			{
				Statement: `    rec		record;`,
			},
			{
				Statement: `begin
    mytype := substr(myname, 1, 2);`,
			},
			{
				Statement: `    if mytype = ''PS'' then
        select into rec * from PSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.backlink = blname then
	    update PSlot set backlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''WS'' then
        select into rec * from WSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.backlink = blname then
	    update WSlot set backlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''PL'' then
        select into rec * from PLine where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.backlink = blname then
	    update PLine set backlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end
' language plpgsql;`,
			},
			{
				Statement: `create function tg_slotlink_a() returns trigger as '
declare
    dummy	integer;`,
			},
			{
				Statement: `begin
    if tg_op = ''INSERT'' then
        if new.slotlink != '''' then
	    dummy := tg_slotlink_set(new.slotlink, new.slotname);`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''UPDATE'' then
        if new.slotlink != old.slotlink then
	    if old.slotlink != '''' then
	        dummy := tg_slotlink_unset(old.slotlink, old.slotname);`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	    if new.slotlink != '''' then
	        dummy := tg_slotlink_set(new.slotlink, new.slotname);`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	else
	    if new.slotname != old.slotname and new.slotlink != '''' then
	        dummy := tg_slotlink_set(new.slotlink, new.slotname);`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if tg_op = ''DELETE'' then
        if old.slotlink != '''' then
	    dummy := tg_slotlink_unset(old.slotlink, old.slotname);`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return old;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create trigger tg_slotlink_a after insert or update or delete
    on PSlot for each row execute procedure tg_slotlink_a('PS');`,
			},
			{
				Statement: `create trigger tg_slotlink_a after insert or update or delete
    on WSlot for each row execute procedure tg_slotlink_a('WS');`,
			},
			{
				Statement: `create trigger tg_slotlink_a after insert or update or delete
    on IFace for each row execute procedure tg_slotlink_a('IF');`,
			},
			{
				Statement: `create trigger tg_slotlink_a after insert or update or delete
    on HSlot for each row execute procedure tg_slotlink_a('HS');`,
			},
			{
				Statement: `create trigger tg_slotlink_a after insert or update or delete
    on PHone for each row execute procedure tg_slotlink_a('PH');`,
			},
			{
				Statement: `create function tg_slotlink_set(bpchar, bpchar)
returns integer as '
declare
    myname	alias for $1;`,
			},
			{
				Statement: `    blname	alias for $2;`,
			},
			{
				Statement: `    mytype	char(2);`,
			},
			{
				Statement: `    link	char(4);`,
			},
			{
				Statement: `    rec		record;`,
			},
			{
				Statement: `begin
    mytype := substr(myname, 1, 2);`,
			},
			{
				Statement: `    link := mytype || substr(blname, 1, 2);`,
			},
			{
				Statement: `    if link = ''PHPH'' then
        raise exception
		''slotlink between two phones does not make sense'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if link in (''PHHS'', ''HSPH'') then
        raise exception
		''link of phone to hub does not make sense'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if link in (''PHIF'', ''IFPH'') then
        raise exception
		''link of phone to hub does not make sense'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if link in (''PSWS'', ''WSPS'') then
        raise exception
		''slotlink from patchslot to wallslot not permitted'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''PS'' then
        select into rec * from PSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink != blname then
	    update PSlot set slotlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''WS'' then
        select into rec * from WSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink != blname then
	    update WSlot set slotlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''IF'' then
        select into rec * from IFace where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink != blname then
	    update IFace set slotlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''HS'' then
        select into rec * from HSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink != blname then
	    update HSlot set slotlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''PH'' then
        select into rec * from PHone where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    raise exception ''% does not exist'', myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink != blname then
	    update PHone set slotlink = blname where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    raise exception ''illegal slotlink beginning with %'', mytype;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create function tg_slotlink_unset(bpchar, bpchar)
returns integer as '
declare
    myname	alias for $1;`,
			},
			{
				Statement: `    blname	alias for $2;`,
			},
			{
				Statement: `    mytype	char(2);`,
			},
			{
				Statement: `    rec		record;`,
			},
			{
				Statement: `begin
    mytype := substr(myname, 1, 2);`,
			},
			{
				Statement: `    if mytype = ''PS'' then
        select into rec * from PSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink = blname then
	    update PSlot set slotlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''WS'' then
        select into rec * from WSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink = blname then
	    update WSlot set slotlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''IF'' then
        select into rec * from IFace where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink = blname then
	    update IFace set slotlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''HS'' then
        select into rec * from HSlot where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink = blname then
	    update HSlot set slotlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if mytype = ''PH'' then
        select into rec * from PHone where slotname = myname;`,
			},
			{
				Statement: `	if not found then
	    return 0;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if rec.slotlink = blname then
	    update PHone set slotlink = '''' where slotname = myname;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return 0;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create function pslot_backlink_view(bpchar)
returns text as '
<<outer>>
declare
    rec		record;`,
			},
			{
				Statement: `    bltype	char(2);`,
			},
			{
				Statement: `    retval	text;`,
			},
			{
				Statement: `begin
    select into rec * from PSlot where slotname = $1;`,
			},
			{
				Statement: `    if not found then
        return '''';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if rec.backlink = '''' then
        return ''-'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    bltype := substr(rec.backlink, 1, 2);`,
			},
			{
				Statement: `    if bltype = ''PL'' then
        declare
	    rec		record;`,
			},
			{
				Statement: `	begin
	    select into rec * from PLine where slotname = "outer".rec.backlink;`,
			},
			{
				Statement: `	    retval := ''Phone line '' || trim(rec.phonenumber);`,
			},
			{
				Statement: `	    if rec.comment != '''' then
	        retval := retval || '' ('';`,
			},
			{
				Statement: `		retval := retval || rec.comment;`,
			},
			{
				Statement: `		retval := retval || '')'';`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	    return retval;`,
			},
			{
				Statement: `	end;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if bltype = ''WS'' then
        select into rec * from WSlot where slotname = rec.backlink;`,
			},
			{
				Statement: `	retval := trim(rec.slotname) || '' in room '';`,
			},
			{
				Statement: `	retval := retval || trim(rec.roomno);`,
			},
			{
				Statement: `	retval := retval || '' -> '';`,
			},
			{
				Statement: `	return retval || wslot_slotlink_view(rec.slotname);`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return rec.backlink;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create function pslot_slotlink_view(bpchar)
returns text as '
declare
    psrec	record;`,
			},
			{
				Statement: `    sltype	char(2);`,
			},
			{
				Statement: `    retval	text;`,
			},
			{
				Statement: `begin
    select into psrec * from PSlot where slotname = $1;`,
			},
			{
				Statement: `    if not found then
        return '''';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if psrec.slotlink = '''' then
        return ''-'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    sltype := substr(psrec.slotlink, 1, 2);`,
			},
			{
				Statement: `    if sltype = ''PS'' then
	retval := trim(psrec.slotlink) || '' -> '';`,
			},
			{
				Statement: `	return retval || pslot_backlink_view(psrec.slotlink);`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if sltype = ''HS'' then
        retval := comment from Hub H, HSlot HS
			where HS.slotname = psrec.slotlink
			  and H.name = HS.hubname;`,
			},
			{
				Statement: `        retval := retval || '' slot '';`,
			},
			{
				Statement: `	retval := retval || slotno::text from HSlot
			where slotname = psrec.slotlink;`,
			},
			{
				Statement: `	return retval;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return psrec.slotlink;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create function wslot_slotlink_view(bpchar)
returns text as '
declare
    rec		record;`,
			},
			{
				Statement: `    sltype	char(2);`,
			},
			{
				Statement: `    retval	text;`,
			},
			{
				Statement: `begin
    select into rec * from WSlot where slotname = $1;`,
			},
			{
				Statement: `    if not found then
        return '''';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if rec.slotlink = '''' then
        return ''-'';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    sltype := substr(rec.slotlink, 1, 2);`,
			},
			{
				Statement: `    if sltype = ''PH'' then
        select into rec * from PHone where slotname = rec.slotlink;`,
			},
			{
				Statement: `	retval := ''Phone '' || trim(rec.slotname);`,
			},
			{
				Statement: `	if rec.comment != '''' then
	    retval := retval || '' ('';`,
			},
			{
				Statement: `	    retval := retval || rec.comment;`,
			},
			{
				Statement: `	    retval := retval || '')'';`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return retval;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if sltype = ''IF'' then
	declare
	    syrow	System%RowType;`,
			},
			{
				Statement: `	    ifrow	IFace%ROWTYPE;`,
			},
			{
				Statement: `        begin
	    select into ifrow * from IFace where slotname = rec.slotlink;`,
			},
			{
				Statement: `	    select into syrow * from System where name = ifrow.sysname;`,
			},
			{
				Statement: `	    retval := syrow.name || '' IF '';`,
			},
			{
				Statement: `	    retval := retval || ifrow.ifname;`,
			},
			{
				Statement: `	    if syrow.comment != '''' then
	        retval := retval || '' ('';`,
			},
			{
				Statement: `		retval := retval || syrow.comment;`,
			},
			{
				Statement: `		retval := retval || '')'';`,
			},
			{
				Statement: `	    end if;`,
			},
			{
				Statement: `	    return retval;`,
			},
			{
				Statement: `	end;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return rec.slotlink;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `' language plpgsql;`,
			},
			{
				Statement: `create view Pfield_v1 as select PF.pfname, PF.slotname,
	pslot_backlink_view(PF.slotname) as backside,
	pslot_slotlink_view(PF.slotname) as patch
    from PSlot PF;`,
			},
			{
				Statement: `insert into Room values ('001', 'Entrance');`,
			},
			{
				Statement: `insert into Room values ('002', 'Office');`,
			},
			{
				Statement: `insert into Room values ('003', 'Office');`,
			},
			{
				Statement: `insert into Room values ('004', 'Technical');`,
			},
			{
				Statement: `insert into Room values ('101', 'Office');`,
			},
			{
				Statement: `insert into Room values ('102', 'Conference');`,
			},
			{
				Statement: `insert into Room values ('103', 'Restroom');`,
			},
			{
				Statement: `insert into Room values ('104', 'Technical');`,
			},
			{
				Statement: `insert into Room values ('105', 'Office');`,
			},
			{
				Statement: `insert into Room values ('106', 'Office');`,
			},
			{
				Statement: `insert into WSlot values ('WS.001.1a', '001', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.001.1b', '001', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.001.2a', '001', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.001.2b', '001', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.001.3a', '001', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.001.3b', '001', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.002.1a', '002', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.002.1b', '002', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.002.2a', '002', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.002.2b', '002', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.002.3a', '002', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.002.3b', '002', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.003.1a', '003', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.003.1b', '003', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.003.2a', '003', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.003.2b', '003', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.003.3a', '003', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.003.3b', '003', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.101.1a', '101', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.101.1b', '101', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.101.2a', '101', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.101.2b', '101', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.101.3a', '101', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.101.3b', '101', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.102.1a', '102', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.102.1b', '102', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.102.2a', '102', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.102.2b', '102', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.102.3a', '102', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.102.3b', '102', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.105.1a', '105', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.105.1b', '105', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.105.2a', '105', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.105.2b', '105', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.105.3a', '105', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.105.3b', '105', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.106.1a', '106', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.106.1b', '106', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.106.2a', '106', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.106.2b', '106', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.106.3a', '106', '', '');`,
			},
			{
				Statement: `insert into WSlot values ('WS.106.3b', '106', '', '');`,
			},
			{
				Statement: `insert into PField values ('PF0_1', 'Wallslots basement');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.a1', 'PF0_1', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.a2', 'PF0_1', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.a3', 'PF0_1', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.a4', 'PF0_1', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.a5', 'PF0_1', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.a6', 'PF0_1', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.b1', 'PF0_1', '', 'WS.002.1a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.b2', 'PF0_1', '', 'WS.002.1b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.b3', 'PF0_1', '', 'WS.002.2a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.b4', 'PF0_1', '', 'WS.002.2b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.b5', 'PF0_1', '', 'WS.002.3a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.b6', 'PF0_1', '', 'WS.002.3b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.c1', 'PF0_1', '', 'WS.003.1a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.c2', 'PF0_1', '', 'WS.003.1b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.c3', 'PF0_1', '', 'WS.003.2a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.c4', 'PF0_1', '', 'WS.003.2b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.c5', 'PF0_1', '', 'WS.003.3a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.c6', 'PF0_1', '', 'WS.003.3b');`,
			},
			{
				Statement: `insert into PField values ('PF0_X', 'Phonelines basement');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.ta1', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.ta2', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.ta3', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.ta4', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.ta5', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.ta6', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.tb1', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.tb2', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.tb3', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.tb4', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.tb5', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.base.tb6', 'PF0_X', '', '');`,
			},
			{
				Statement: `insert into PField values ('PF1_1', 'Wallslots first floor');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.a1', 'PF1_1', '', 'WS.101.1a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.a2', 'PF1_1', '', 'WS.101.1b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.a3', 'PF1_1', '', 'WS.101.2a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.a4', 'PF1_1', '', 'WS.101.2b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.a5', 'PF1_1', '', 'WS.101.3a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.a6', 'PF1_1', '', 'WS.101.3b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.b1', 'PF1_1', '', 'WS.102.1a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.b2', 'PF1_1', '', 'WS.102.1b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.b3', 'PF1_1', '', 'WS.102.2a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.b4', 'PF1_1', '', 'WS.102.2b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.b5', 'PF1_1', '', 'WS.102.3a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.b6', 'PF1_1', '', 'WS.102.3b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.c1', 'PF1_1', '', 'WS.105.1a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.c2', 'PF1_1', '', 'WS.105.1b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.c3', 'PF1_1', '', 'WS.105.2a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.c4', 'PF1_1', '', 'WS.105.2b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.c5', 'PF1_1', '', 'WS.105.3a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.c6', 'PF1_1', '', 'WS.105.3b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.d1', 'PF1_1', '', 'WS.106.1a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.d2', 'PF1_1', '', 'WS.106.1b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.d3', 'PF1_1', '', 'WS.106.2a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.d4', 'PF1_1', '', 'WS.106.2b');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.d5', 'PF1_1', '', 'WS.106.3a');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.d6', 'PF1_1', '', 'WS.106.3b');`,
			},
			{
				Statement: `update PSlot set backlink = 'WS.001.1a' where slotname = 'PS.base.a1';`,
			},
			{
				Statement: `update PSlot set backlink = 'WS.001.1b' where slotname = 'PS.base.a3';`,
			},
			{
				Statement: `select * from WSlot where roomno = '001' order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, `PS.base.a3`}, {`WS.001.2a`, 001, ``, ``}, {`WS.001.2b`, 001, ``, ``}, {`WS.001.3a`, 001, ``, ``}, {`WS.001.3b`, 001, ``, ``}},
			},
			{
				Statement: `select * from PSlot where slotname ~ 'PS.base.a' order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, ``}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.1b`}, {`PS.base.a4`, `PF0_1`, ``, ``}, {`PS.base.a5`, `PF0_1`, ``, ``}, {`PS.base.a6`, `PF0_1`, ``, ``}},
			},
			{
				Statement: `update PSlot set backlink = 'WS.001.2a' where slotname = 'PS.base.a3';`,
			},
			{
				Statement: `select * from WSlot where roomno = '001' order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, ``}, {`WS.001.2a`, 001, ``, `PS.base.a3`}, {`WS.001.2b`, 001, ``, ``}, {`WS.001.3a`, 001, ``, ``}, {`WS.001.3b`, 001, ``, ``}},
			},
			{
				Statement: `select * from PSlot where slotname ~ 'PS.base.a' order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, ``}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.2a`}, {`PS.base.a4`, `PF0_1`, ``, ``}, {`PS.base.a5`, `PF0_1`, ``, ``}, {`PS.base.a6`, `PF0_1`, ``, ``}},
			},
			{
				Statement: `update PSlot set backlink = 'WS.001.1b' where slotname = 'PS.base.a2';`,
			},
			{
				Statement: `select * from WSlot where roomno = '001' order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, `PS.base.a2`}, {`WS.001.2a`, 001, ``, `PS.base.a3`}, {`WS.001.2b`, 001, ``, ``}, {`WS.001.3a`, 001, ``, ``}, {`WS.001.3b`, 001, ``, ``}},
			},
			{
				Statement: `select * from PSlot where slotname ~ 'PS.base.a' order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, `WS.001.1b`}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.2a`}, {`PS.base.a4`, `PF0_1`, ``, ``}, {`PS.base.a5`, `PF0_1`, ``, ``}, {`PS.base.a6`, `PF0_1`, ``, ``}},
			},
			{
				Statement: `update WSlot set backlink = 'PS.base.a4' where slotname = 'WS.001.2b';`,
			},
			{
				Statement: `update WSlot set backlink = 'PS.base.a6' where slotname = 'WS.001.3a';`,
			},
			{
				Statement: `select * from WSlot where roomno = '001' order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, `PS.base.a2`}, {`WS.001.2a`, 001, ``, `PS.base.a3`}, {`WS.001.2b`, 001, ``, `PS.base.a4`}, {`WS.001.3a`, 001, ``, `PS.base.a6`}, {`WS.001.3b`, 001, ``, ``}},
			},
			{
				Statement: `select * from PSlot where slotname ~ 'PS.base.a' order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, `WS.001.1b`}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.2a`}, {`PS.base.a4`, `PF0_1`, ``, `WS.001.2b`}, {`PS.base.a5`, `PF0_1`, ``, ``}, {`PS.base.a6`, `PF0_1`, ``, `WS.001.3a`}},
			},
			{
				Statement: `update WSlot set backlink = 'PS.base.a6' where slotname = 'WS.001.3b';`,
			},
			{
				Statement: `select * from WSlot where roomno = '001' order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, `PS.base.a2`}, {`WS.001.2a`, 001, ``, `PS.base.a3`}, {`WS.001.2b`, 001, ``, `PS.base.a4`}, {`WS.001.3a`, 001, ``, ``}, {`WS.001.3b`, 001, ``, `PS.base.a6`}},
			},
			{
				Statement: `select * from PSlot where slotname ~ 'PS.base.a' order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, `WS.001.1b`}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.2a`}, {`PS.base.a4`, `PF0_1`, ``, `WS.001.2b`}, {`PS.base.a5`, `PF0_1`, ``, ``}, {`PS.base.a6`, `PF0_1`, ``, `WS.001.3b`}},
			},
			{
				Statement: `update WSlot set backlink = 'PS.base.a5' where slotname = 'WS.001.3a';`,
			},
			{
				Statement: `select * from WSlot where roomno = '001' order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, `PS.base.a2`}, {`WS.001.2a`, 001, ``, `PS.base.a3`}, {`WS.001.2b`, 001, ``, `PS.base.a4`}, {`WS.001.3a`, 001, ``, `PS.base.a5`}, {`WS.001.3b`, 001, ``, `PS.base.a6`}},
			},
			{
				Statement: `select * from PSlot where slotname ~ 'PS.base.a' order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, `WS.001.1b`}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.2a`}, {`PS.base.a4`, `PF0_1`, ``, `WS.001.2b`}, {`PS.base.a5`, `PF0_1`, ``, `WS.001.3a`}, {`PS.base.a6`, `PF0_1`, ``, `WS.001.3b`}},
			},
			{
				Statement: `insert into PField values ('PF1_2', 'Phonelines first floor');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.ta1', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.ta2', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.ta3', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.ta4', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.ta5', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.ta6', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.tb1', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.tb2', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.tb3', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.tb4', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.tb5', 'PF1_2', '', '');`,
			},
			{
				Statement: `insert into PSlot values ('PS.first.tb6', 'PF1_2', '', '');`,
			},
			{
				Statement: `update PField set name = 'PF0_2' where name = 'PF0_X';`,
			},
			{
				Statement: `select * from PSlot order by slotname;`,
				Results:   []sql.Row{{`PS.base.a1`, `PF0_1`, ``, `WS.001.1a`}, {`PS.base.a2`, `PF0_1`, ``, `WS.001.1b`}, {`PS.base.a3`, `PF0_1`, ``, `WS.001.2a`}, {`PS.base.a4`, `PF0_1`, ``, `WS.001.2b`}, {`PS.base.a5`, `PF0_1`, ``, `WS.001.3a`}, {`PS.base.a6`, `PF0_1`, ``, `WS.001.3b`}, {`PS.base.b1`, `PF0_1`, ``, `WS.002.1a`}, {`PS.base.b2`, `PF0_1`, ``, `WS.002.1b`}, {`PS.base.b3`, `PF0_1`, ``, `WS.002.2a`}, {`PS.base.b4`, `PF0_1`, ``, `WS.002.2b`}, {`PS.base.b5`, `PF0_1`, ``, `WS.002.3a`}, {`PS.base.b6`, `PF0_1`, ``, `WS.002.3b`}, {`PS.base.c1`, `PF0_1`, ``, `WS.003.1a`}, {`PS.base.c2`, `PF0_1`, ``, `WS.003.1b`}, {`PS.base.c3`, `PF0_1`, ``, `WS.003.2a`}, {`PS.base.c4`, `PF0_1`, ``, `WS.003.2b`}, {`PS.base.c5`, `PF0_1`, ``, `WS.003.3a`}, {`PS.base.c6`, `PF0_1`, ``, `WS.003.3b`}, {`PS.base.ta1`, `PF0_2`, ``, ``}, {`PS.base.ta2`, `PF0_2`, ``, ``}, {`PS.base.ta3`, `PF0_2`, ``, ``}, {`PS.base.ta4`, `PF0_2`, ``, ``}, {`PS.base.ta5`, `PF0_2`, ``, ``}, {`PS.base.ta6`, `PF0_2`, ``, ``}, {`PS.base.tb1`, `PF0_2`, ``, ``}, {`PS.base.tb2`, `PF0_2`, ``, ``}, {`PS.base.tb3`, `PF0_2`, ``, ``}, {`PS.base.tb4`, `PF0_2`, ``, ``}, {`PS.base.tb5`, `PF0_2`, ``, ``}, {`PS.base.tb6`, `PF0_2`, ``, ``}, {`PS.first.a1`, `PF1_1`, ``, `WS.101.1a`}, {`PS.first.a2`, `PF1_1`, ``, `WS.101.1b`}, {`PS.first.a3`, `PF1_1`, ``, `WS.101.2a`}, {`PS.first.a4`, `PF1_1`, ``, `WS.101.2b`}, {`PS.first.a5`, `PF1_1`, ``, `WS.101.3a`}, {`PS.first.a6`, `PF1_1`, ``, `WS.101.3b`}, {`PS.first.b1`, `PF1_1`, ``, `WS.102.1a`}, {`PS.first.b2`, `PF1_1`, ``, `WS.102.1b`}, {`PS.first.b3`, `PF1_1`, ``, `WS.102.2a`}, {`PS.first.b4`, `PF1_1`, ``, `WS.102.2b`}, {`PS.first.b5`, `PF1_1`, ``, `WS.102.3a`}, {`PS.first.b6`, `PF1_1`, ``, `WS.102.3b`}, {`PS.first.c1`, `PF1_1`, ``, `WS.105.1a`}, {`PS.first.c2`, `PF1_1`, ``, `WS.105.1b`}, {`PS.first.c3`, `PF1_1`, ``, `WS.105.2a`}, {`PS.first.c4`, `PF1_1`, ``, `WS.105.2b`}, {`PS.first.c5`, `PF1_1`, ``, `WS.105.3a`}, {`PS.first.c6`, `PF1_1`, ``, `WS.105.3b`}, {`PS.first.d1`, `PF1_1`, ``, `WS.106.1a`}, {`PS.first.d2`, `PF1_1`, ``, `WS.106.1b`}, {`PS.first.d3`, `PF1_1`, ``, `WS.106.2a`}, {`PS.first.d4`, `PF1_1`, ``, `WS.106.2b`}, {`PS.first.d5`, `PF1_1`, ``, `WS.106.3a`}, {`PS.first.d6`, `PF1_1`, ``, `WS.106.3b`}, {`PS.first.ta1`, `PF1_2`, ``, ``}, {`PS.first.ta2`, `PF1_2`, ``, ``}, {`PS.first.ta3`, `PF1_2`, ``, ``}, {`PS.first.ta4`, `PF1_2`, ``, ``}, {`PS.first.ta5`, `PF1_2`, ``, ``}, {`PS.first.ta6`, `PF1_2`, ``, ``}, {`PS.first.tb1`, `PF1_2`, ``, ``}, {`PS.first.tb2`, `PF1_2`, ``, ``}, {`PS.first.tb3`, `PF1_2`, ``, ``}, {`PS.first.tb4`, `PF1_2`, ``, ``}, {`PS.first.tb5`, `PF1_2`, ``, ``}, {`PS.first.tb6`, `PF1_2`, ``, ``}},
			},
			{
				Statement: `select * from WSlot order by slotname;`,
				Results:   []sql.Row{{`WS.001.1a`, 001, ``, `PS.base.a1`}, {`WS.001.1b`, 001, ``, `PS.base.a2`}, {`WS.001.2a`, 001, ``, `PS.base.a3`}, {`WS.001.2b`, 001, ``, `PS.base.a4`}, {`WS.001.3a`, 001, ``, `PS.base.a5`}, {`WS.001.3b`, 001, ``, `PS.base.a6`}, {`WS.002.1a`, 002, ``, `PS.base.b1`}, {`WS.002.1b`, 002, ``, `PS.base.b2`}, {`WS.002.2a`, 002, ``, `PS.base.b3`}, {`WS.002.2b`, 002, ``, `PS.base.b4`}, {`WS.002.3a`, 002, ``, `PS.base.b5`}, {`WS.002.3b`, 002, ``, `PS.base.b6`}, {`WS.003.1a`, 003, ``, `PS.base.c1`}, {`WS.003.1b`, 003, ``, `PS.base.c2`}, {`WS.003.2a`, 003, ``, `PS.base.c3`}, {`WS.003.2b`, 003, ``, `PS.base.c4`}, {`WS.003.3a`, 003, ``, `PS.base.c5`}, {`WS.003.3b`, 003, ``, `PS.base.c6`}, {`WS.101.1a`, 101, ``, `PS.first.a1`}, {`WS.101.1b`, 101, ``, `PS.first.a2`}, {`WS.101.2a`, 101, ``, `PS.first.a3`}, {`WS.101.2b`, 101, ``, `PS.first.a4`}, {`WS.101.3a`, 101, ``, `PS.first.a5`}, {`WS.101.3b`, 101, ``, `PS.first.a6`}, {`WS.102.1a`, 102, ``, `PS.first.b1`}, {`WS.102.1b`, 102, ``, `PS.first.b2`}, {`WS.102.2a`, 102, ``, `PS.first.b3`}, {`WS.102.2b`, 102, ``, `PS.first.b4`}, {`WS.102.3a`, 102, ``, `PS.first.b5`}, {`WS.102.3b`, 102, ``, `PS.first.b6`}, {`WS.105.1a`, 105, ``, `PS.first.c1`}, {`WS.105.1b`, 105, ``, `PS.first.c2`}, {`WS.105.2a`, 105, ``, `PS.first.c3`}, {`WS.105.2b`, 105, ``, `PS.first.c4`}, {`WS.105.3a`, 105, ``, `PS.first.c5`}, {`WS.105.3b`, 105, ``, `PS.first.c6`}, {`WS.106.1a`, 106, ``, `PS.first.d1`}, {`WS.106.1b`, 106, ``, `PS.first.d2`}, {`WS.106.2a`, 106, ``, `PS.first.d3`}, {`WS.106.2b`, 106, ``, `PS.first.d4`}, {`WS.106.3a`, 106, ``, `PS.first.d5`}, {`WS.106.3b`, 106, ``, `PS.first.d6`}},
			},
			{
				Statement: `insert into PLine values ('PL.001', '-0', 'Central call', 'PS.base.ta1');`,
			},
			{
				Statement: `insert into PLine values ('PL.002', '-101', '', 'PS.base.ta2');`,
			},
			{
				Statement: `insert into PLine values ('PL.003', '-102', '', 'PS.base.ta3');`,
			},
			{
				Statement: `insert into PLine values ('PL.004', '-103', '', 'PS.base.ta5');`,
			},
			{
				Statement: `insert into PLine values ('PL.005', '-104', '', 'PS.base.ta6');`,
			},
			{
				Statement: `insert into PLine values ('PL.006', '-106', '', 'PS.base.tb2');`,
			},
			{
				Statement: `insert into PLine values ('PL.007', '-108', '', 'PS.base.tb3');`,
			},
			{
				Statement: `insert into PLine values ('PL.008', '-109', '', 'PS.base.tb4');`,
			},
			{
				Statement: `insert into PLine values ('PL.009', '-121', '', 'PS.base.tb5');`,
			},
			{
				Statement: `insert into PLine values ('PL.010', '-122', '', 'PS.base.tb6');`,
			},
			{
				Statement: `insert into PLine values ('PL.015', '-134', '', 'PS.first.ta1');`,
			},
			{
				Statement: `insert into PLine values ('PL.016', '-137', '', 'PS.first.ta3');`,
			},
			{
				Statement: `insert into PLine values ('PL.017', '-139', '', 'PS.first.ta4');`,
			},
			{
				Statement: `insert into PLine values ('PL.018', '-362', '', 'PS.first.tb1');`,
			},
			{
				Statement: `insert into PLine values ('PL.019', '-363', '', 'PS.first.tb2');`,
			},
			{
				Statement: `insert into PLine values ('PL.020', '-364', '', 'PS.first.tb3');`,
			},
			{
				Statement: `insert into PLine values ('PL.021', '-365', '', 'PS.first.tb5');`,
			},
			{
				Statement: `insert into PLine values ('PL.022', '-367', '', 'PS.first.tb6');`,
			},
			{
				Statement: `insert into PLine values ('PL.028', '-501', 'Fax entrance', 'PS.base.ta2');`,
			},
			{
				Statement: `insert into PLine values ('PL.029', '-502', 'Fax first floor', 'PS.first.ta1');`,
			},
			{
				Statement: `insert into PHone values ('PH.hc001', 'Hicom standard', 'WS.001.1a');`,
			},
			{
				Statement: `update PSlot set slotlink = 'PS.base.ta1' where slotname = 'PS.base.a1';`,
			},
			{
				Statement: `insert into PHone values ('PH.hc002', 'Hicom standard', 'WS.002.1a');`,
			},
			{
				Statement: `update PSlot set slotlink = 'PS.base.ta5' where slotname = 'PS.base.b1';`,
			},
			{
				Statement: `insert into PHone values ('PH.hc003', 'Hicom standard', 'WS.002.2a');`,
			},
			{
				Statement: `update PSlot set slotlink = 'PS.base.tb2' where slotname = 'PS.base.b3';`,
			},
			{
				Statement: `insert into PHone values ('PH.fax001', 'Canon fax', 'WS.001.2a');`,
			},
			{
				Statement: `update PSlot set slotlink = 'PS.base.ta2' where slotname = 'PS.base.a3';`,
			},
			{
				Statement: `insert into Hub values ('base.hub1', 'Patchfield PF0_1 hub', 16);`,
			},
			{
				Statement: `insert into System values ('orion', 'PC');`,
			},
			{
				Statement: `insert into IFace values ('IF', 'orion', 'eth0', 'WS.002.1b');`,
			},
			{
				Statement: `update PSlot set slotlink = 'HS.base.hub1.1' where slotname = 'PS.base.b2';`,
			},
			{
				Statement: `select * from PField_v1 where pfname = 'PF0_1' order by slotname;`,
				Results:   []sql.Row{{`PF0_1`, `PS.base.a1`, `WS.001.1a in room 001 -> Phone PH.hc001 (Hicom standard)`, `PS.base.ta1 -> Phone line -0 (Central call)`}, {`PF0_1`, `PS.base.a2`, `WS.001.1b in room 001 -> -`, `-`}, {`PF0_1`, `PS.base.a3`, `WS.001.2a in room 001 -> Phone PH.fax001 (Canon fax)`, `PS.base.ta2 -> Phone line -501 (Fax entrance)`}, {`PF0_1`, `PS.base.a4`, `WS.001.2b in room 001 -> -`, `-`}, {`PF0_1`, `PS.base.a5`, `WS.001.3a in room 001 -> -`, `-`}, {`PF0_1`, `PS.base.a6`, `WS.001.3b in room 001 -> -`, `-`}, {`PF0_1`, `PS.base.b1`, `WS.002.1a in room 002 -> Phone PH.hc002 (Hicom standard)`, `PS.base.ta5 -> Phone line -103`}, {`PF0_1`, `PS.base.b2`, `WS.002.1b in room 002 -> orion IF eth0 (PC)`, `Patchfield PF0_1 hub slot 1`}, {`PF0_1`, `PS.base.b3`, `WS.002.2a in room 002 -> Phone PH.hc003 (Hicom standard)`, `PS.base.tb2 -> Phone line -106`}, {`PF0_1`, `PS.base.b4`, `WS.002.2b in room 002 -> -`, `-`}, {`PF0_1`, `PS.base.b5`, `WS.002.3a in room 002 -> -`, `-`}, {`PF0_1`, `PS.base.b6`, `WS.002.3b in room 002 -> -`, `-`}, {`PF0_1`, `PS.base.c1`, `WS.003.1a in room 003 -> -`, `-`}, {`PF0_1`, `PS.base.c2`, `WS.003.1b in room 003 -> -`, `-`}, {`PF0_1`, `PS.base.c3`, `WS.003.2a in room 003 -> -`, `-`}, {`PF0_1`, `PS.base.c4`, `WS.003.2b in room 003 -> -`, `-`}, {`PF0_1`, `PS.base.c5`, `WS.003.3a in room 003 -> -`, `-`}, {`PF0_1`, `PS.base.c6`, `WS.003.3b in room 003 -> -`, `-`}},
			},
			{
				Statement: `select * from PField_v1 where pfname = 'PF0_2' order by slotname;`,
				Results:   []sql.Row{{`PF0_2`, `PS.base.ta1`, `Phone line -0 (Central call)`, `PS.base.a1 -> WS.001.1a in room 001 -> Phone PH.hc001 (Hicom standard)`}, {`PF0_2`, `PS.base.ta2`, `Phone line -501 (Fax entrance)`, `PS.base.a3 -> WS.001.2a in room 001 -> Phone PH.fax001 (Canon fax)`}, {`PF0_2`, `PS.base.ta3`, `Phone line -102`, `-`}, {`PF0_2`, `PS.base.ta4`, `-`, `-`}, {`PF0_2`, `PS.base.ta5`, `Phone line -103`, `PS.base.b1 -> WS.002.1a in room 002 -> Phone PH.hc002 (Hicom standard)`}, {`PF0_2`, `PS.base.ta6`, `Phone line -104`, `-`}, {`PF0_2`, `PS.base.tb1`, `-`, `-`}, {`PF0_2`, `PS.base.tb2`, `Phone line -106`, `PS.base.b3 -> WS.002.2a in room 002 -> Phone PH.hc003 (Hicom standard)`}, {`PF0_2`, `PS.base.tb3`, `Phone line -108`, `-`}, {`PF0_2`, `PS.base.tb4`, `Phone line -109`, `-`}, {`PF0_2`, `PS.base.tb5`, `Phone line -121`, `-`}, {`PF0_2`, `PS.base.tb6`, `Phone line -122`, `-`}},
			},
			{
				Statement:   `insert into PField values ('PF1_1', 'should fail due to unique index');`,
				ErrorString: `duplicate key value violates unique constraint "pfield_name"`,
			},
			{
				Statement:   `update PSlot set backlink = 'WS.not.there' where slotname = 'PS.base.a1';`,
				ErrorString: `WS.not.there         does not exist`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_backlink_set(character,character) line 30 at RAISE
PL/pgSQL function tg_backlink_a() line 17 at assignment
update PSlot set backlink = 'XX.illegal' where slotname = 'PS.base.a1';`,
				ErrorString: `illegal backlink beginning with XX`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_backlink_set(character,character) line 47 at RAISE
PL/pgSQL function tg_backlink_a() line 17 at assignment
update PSlot set slotlink = 'PS.not.there' where slotname = 'PS.base.a1';`,
				ErrorString: `PS.not.there         does not exist`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_slotlink_set(character,character) line 30 at RAISE
PL/pgSQL function tg_slotlink_a() line 17 at assignment
update PSlot set slotlink = 'XX.illegal' where slotname = 'PS.base.a1';`,
				ErrorString: `illegal slotlink beginning with XX`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_slotlink_set(character,character) line 77 at RAISE
PL/pgSQL function tg_slotlink_a() line 17 at assignment
insert into HSlot values ('HS', 'base.hub1', 1, '');`,
				ErrorString: `duplicate key value violates unique constraint "hslot_name"`,
			},
			{
				Statement:   `insert into HSlot values ('HS', 'base.hub1', 20, '');`,
				ErrorString: `no manual manipulation of HSlot`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_hslot_biu() line 12 at RAISE
delete from HSlot;`,
				ErrorString: `no manual manipulation of HSlot`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_hslot_bd() line 12 at RAISE
insert into IFace values ('IF', 'notthere', 'eth0', '');`,
				ErrorString: `system "notthere" does not exist`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_iface_biu() line 8 at RAISE
insert into IFace values ('IF', 'orion', 'ethernet_interface_name_too_long', '');`,
				ErrorString: `IFace slotname "IF.orion.ethernet_interface_name_too_long" too long (20 char max)`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tg_iface_biu() line 14 at RAISE
CREATE FUNCTION recursion_test(int,int) RETURNS text AS '
DECLARE rslt text;`,
			},
			{
				Statement: `BEGIN
    IF $1 <= 0 THEN
        rslt = CAST($2 AS TEXT);`,
			},
			{
				Statement: `    ELSE
        rslt = CAST($1 AS TEXT) || '','' || recursion_test($1 - 1, $2);`,
			},
			{
				Statement: `    END IF;`,
			},
			{
				Statement: `    RETURN rslt;`,
			},
			{
				Statement: `END;' LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT recursion_test(4,3);`,
				Results:   []sql.Row{{`4,3,2,1,3`}},
			},
			{
				Statement: `CREATE TABLE found_test_tbl (a int);`,
			},
			{
				Statement: `create function test_found()
  returns boolean as '
  declare
  begin
  insert into found_test_tbl values (1);`,
			},
			{
				Statement: `  if FOUND then
     insert into found_test_tbl values (2);`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  update found_test_tbl set a = 100 where a = 1;`,
			},
			{
				Statement: `  if FOUND then
    insert into found_test_tbl values (3);`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  delete from found_test_tbl where a = 9999; -- matches no rows`,
			},
			{
				Statement: `  if not FOUND then
    insert into found_test_tbl values (4);`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  for i in 1 .. 10 loop
    -- no need to do anything
  end loop;`,
			},
			{
				Statement: `  if FOUND then
    insert into found_test_tbl values (5);`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  -- never executes the loop
  for i in 2 .. 1 loop
    -- no need to do anything
  end loop;`,
			},
			{
				Statement: `  if not FOUND then
    insert into found_test_tbl values (6);`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return true;`,
			},
			{
				Statement: `  end;' language plpgsql;`,
			},
			{
				Statement: `select test_found();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select * from found_test_tbl;`,
				Results:   []sql.Row{{2}, {100}, {3}, {4}, {5}, {6}},
			},
			{
				Statement: `create function test_table_func_rec() returns setof found_test_tbl as '
DECLARE
	rec RECORD;`,
			},
			{
				Statement: `BEGIN
	FOR rec IN select * from found_test_tbl LOOP
		RETURN NEXT rec;`,
			},
			{
				Statement: `	END LOOP;`,
			},
			{
				Statement: `	RETURN;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `select * from test_table_func_rec();`,
				Results:   []sql.Row{{2}, {100}, {3}, {4}, {5}, {6}},
			},
			{
				Statement: `create function test_table_func_row() returns setof found_test_tbl as '
DECLARE
	row found_test_tbl%ROWTYPE;`,
			},
			{
				Statement: `BEGIN
	FOR row IN select * from found_test_tbl LOOP
		RETURN NEXT row;`,
			},
			{
				Statement: `	END LOOP;`,
			},
			{
				Statement: `	RETURN;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `select * from test_table_func_row();`,
				Results:   []sql.Row{{2}, {100}, {3}, {4}, {5}, {6}},
			},
			{
				Statement: `create function test_ret_set_scalar(int,int) returns setof int as '
DECLARE
	i int;`,
			},
			{
				Statement: `BEGIN
	FOR i IN $1 .. $2 LOOP
		RETURN NEXT i + 1;`,
			},
			{
				Statement: `	END LOOP;`,
			},
			{
				Statement: `	RETURN;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `select * from test_ret_set_scalar(1,10);`,
				Results:   []sql.Row{{2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}},
			},
			{
				Statement: `create function test_ret_set_rec_dyn(int) returns setof record as '
DECLARE
	retval RECORD;`,
			},
			{
				Statement: `BEGIN
	IF $1 > 10 THEN
		SELECT INTO retval 5, 10, 15;`,
			},
			{
				Statement: `		RETURN NEXT retval;`,
			},
			{
				Statement: `		RETURN NEXT retval;`,
			},
			{
				Statement: `	ELSE
		SELECT INTO retval 50, 5::numeric, ''xxx''::text;`,
			},
			{
				Statement: `		RETURN NEXT retval;`,
			},
			{
				Statement: `		RETURN NEXT retval;`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `	RETURN;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `SELECT * FROM test_ret_set_rec_dyn(1500) AS (a int, b int, c int);`,
				Results:   []sql.Row{{5, 10, 15}, {5, 10, 15}},
			},
			{
				Statement: `SELECT * FROM test_ret_set_rec_dyn(5) AS (a int, b numeric, c text);`,
				Results:   []sql.Row{{50, 5, `xxx`}, {50, 5, `xxx`}},
			},
			{
				Statement: `create function test_ret_rec_dyn(int) returns record as '
DECLARE
	retval RECORD;`,
			},
			{
				Statement: `BEGIN
	IF $1 > 10 THEN
		SELECT INTO retval 5, 10, 15;`,
			},
			{
				Statement: `		RETURN retval;`,
			},
			{
				Statement: `	ELSE
		SELECT INTO retval 50, 5::numeric, ''xxx''::text;`,
			},
			{
				Statement: `		RETURN retval;`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `SELECT * FROM test_ret_rec_dyn(1500) AS (a int, b int, c int);`,
				Results:   []sql.Row{{5, 10, 15}},
			},
			{
				Statement: `SELECT * FROM test_ret_rec_dyn(5) AS (a int, b numeric, c text);`,
				Results:   []sql.Row{{50, 5, `xxx`}},
			},
			{
				Statement: `create function f1(x anyelement) returns anyelement as $$
begin
  return x + 1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(42) as int, f1(4.5) as num;`,
				Results:   []sql.Row{{43, 5.5}},
			},
			{
				Statement:   `select f1(point(3,4));  -- fail for lack of + operator`,
				ErrorString: `operator does not exist: point + integer`,
			},
			{
				Statement: `QUERY:  x + 1
CONTEXT:  PL/pgSQL function f1(anyelement) line 3 at RETURN
drop function f1(x anyelement);`,
			},
			{
				Statement: `create function f1(x anyelement) returns anyarray as $$
begin
  return array[x + 1, x + 2];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(42) as int, f1(4.5) as num;`,
				Results:   []sql.Row{{`{43,44}`, `{5.5,6.5}`}},
			},
			{
				Statement: `drop function f1(x anyelement);`,
			},
			{
				Statement: `create function f1(x anyarray) returns anyelement as $$
begin
  return x[1];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(array[2,4]) as int, f1(array[4.5, 7.7]) as num;`,
				Results:   []sql.Row{{2, 4.5}},
			},
			{
				Statement:   `select f1(stavalues1) from pg_statistic;  -- fail, can't infer element type`,
				ErrorString: `cannot determine element type of "anyarray" argument`,
			},
			{
				Statement: `drop function f1(x anyarray);`,
			},
			{
				Statement: `create function f1(x anyarray) returns anyarray as $$
begin
  return x;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(array[2,4]) as int, f1(array[4.5, 7.7]) as num;`,
				Results:   []sql.Row{{`{2,4}`, `{4.5,7.7}`}},
			},
			{
				Statement:   `select f1(stavalues1) from pg_statistic;  -- fail, can't infer element type`,
				ErrorString: `PL/pgSQL functions cannot accept type anyarray`,
			},
			{
				Statement: `CONTEXT:  compilation of PL/pgSQL function "f1" near line 1
drop function f1(x anyarray);`,
			},
			{
				Statement: `create function f1(x anyelement) returns anyrange as $$
begin
  return array[x + 1, x + 2];`,
			},
			{
				Statement:   `end$$ language plpgsql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function f1(x anyrange) returns anyarray as $$
begin
  return array[lower(x), upper(x)];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(int4range(42, 49)) as int, f1(float8range(4.5, 7.8)) as num;`,
				Results:   []sql.Row{{`{42,49}`, `{4.5,7.8}`}},
			},
			{
				Statement: `drop function f1(x anyrange);`,
			},
			{
				Statement: `create function f1(x anycompatible, y anycompatible) returns anycompatiblearray as $$
begin
  return array[x, y];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(2, 4) as int, f1(2, 4.5) as num;`,
				Results:   []sql.Row{{`{2,4}`, `{2,4.5}`}},
			},
			{
				Statement: `drop function f1(x anycompatible, y anycompatible);`,
			},
			{
				Statement: `create function f1(x anycompatiblerange, y anycompatible, z anycompatible) returns anycompatiblearray as $$
begin
  return array[lower(x), upper(x), y, z];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(int4range(42, 49), 11, 2::smallint) as int, f1(float8range(4.5, 7.8), 7.8, 11::real) as num;`,
				Results:   []sql.Row{{`{42,49,11,2}`, `{4.5,7.8,7.8,11}`}},
			},
			{
				Statement:   `select f1(int4range(42, 49), 11, 4.5) as fail;  -- range type doesn't fit`,
				ErrorString: `function f1(int4range, integer, numeric) does not exist`,
			},
			{
				Statement: `drop function f1(x anycompatiblerange, y anycompatible, z anycompatible);`,
			},
			{
				Statement: `create function f1(x anycompatible) returns anycompatiblerange as $$
begin
  return array[x + 1, x + 2];`,
			},
			{
				Statement:   `end$$ language plpgsql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function f1(x anycompatiblerange, y anycompatiblearray) returns anycompatiblerange as $$
begin
  return x;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(int4range(42, 49), array[11]) as int, f1(float8range(4.5, 7.8), array[7]) as num;`,
				Results:   []sql.Row{{`[42,49)`, `[4.5,7.8)`}},
			},
			{
				Statement: `drop function f1(x anycompatiblerange, y anycompatiblearray);`,
			},
			{
				Statement: `create function f1(a anyelement, b anyarray,
                   c anycompatible, d anycompatible,
                   OUT x anyarray, OUT y anycompatiblearray)
as $$
begin
  x := a || b;`,
			},
			{
				Statement: `  y := array[c, d];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from f1(11, array[1, 2], 42, 34.5);`,
				Results: []sql.Row{{`{11,1,2}`, `integer[]`, `{42,34.5}`, `numeric[]`}},
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from f1(11, array[1, 2], point(1,2), point(3,4));`,
				Results: []sql.Row{{`{11,1,2}`, `integer[]`, `{"(1,2)","(3,4)"}`, `point[]`}},
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from f1(11, '{1,2}', point(1,2), '(3,4)');`,
				Results: []sql.Row{{`{11,1,2}`, `integer[]`, `{"(1,2)","(3,4)"}`, `point[]`}},
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from f1(11, array[1, 2.2], 42, 34.5);  -- fail`,
				ErrorString: `function f1(integer, numeric[], integer, numeric) does not exist`,
			},
			{
				Statement: `drop function f1(a anyelement, b anyarray,
                 c anycompatible, d anycompatible);`,
			},
			{
				Statement: `create function f1(in i int, out j int) returns int as $$
begin
  return i+1;`,
			},
			{
				Statement:   `end$$ language plpgsql;`,
				ErrorString: `RETURN cannot have a parameter in function with OUT parameters`,
			},
			{
				Statement: `create function f1(in i int, out j int) as $$
begin
  j := i+1;`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(42);`,
				Results:   []sql.Row{{43}},
			},
			{
				Statement: `select * from f1(42);`,
				Results:   []sql.Row{{43}},
			},
			{
				Statement: `create or replace function f1(inout i int) as $$
begin
  i := i+1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(42);`,
				Results:   []sql.Row{{43}},
			},
			{
				Statement: `select * from f1(42);`,
				Results:   []sql.Row{{43}},
			},
			{
				Statement: `drop function f1(int);`,
			},
			{
				Statement: `create function f1(in i int, out j int) returns setof int as $$
begin
  j := i+1;`,
			},
			{
				Statement: `  return next;`,
			},
			{
				Statement: `  j := i+2;`,
			},
			{
				Statement: `  return next;`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select * from f1(42);`,
				Results:   []sql.Row{{43}, {44}},
			},
			{
				Statement: `drop function f1(int);`,
			},
			{
				Statement: `create function f1(in i int, out j int, out k text) as $$
begin
  j := i;`,
			},
			{
				Statement: `  j := j+1;`,
			},
			{
				Statement: `  k := 'foo';`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select f1(42);`,
				Results:   []sql.Row{{`(43,foo)`}},
			},
			{
				Statement: `select * from f1(42);`,
				Results:   []sql.Row{{43, `foo`}},
			},
			{
				Statement: `drop function f1(int);`,
			},
			{
				Statement: `create function f1(in i int, out j int, out k text) returns setof record as $$
begin
  j := i+1;`,
			},
			{
				Statement: `  k := 'foo';`,
			},
			{
				Statement: `  return next;`,
			},
			{
				Statement: `  j := j+1;`,
			},
			{
				Statement: `  k := 'foot';`,
			},
			{
				Statement: `  return next;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select * from f1(42);`,
				Results:   []sql.Row{{43, `foo`}, {44, `foot`}},
			},
			{
				Statement: `drop function f1(int);`,
			},
			{
				Statement: `create function duplic(in i anyelement, out j anyelement, out k anyarray) as $$
begin
  j := i;`,
			},
			{
				Statement: `  k := array[j,j];`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select * from duplic(42);`,
				Results:   []sql.Row{{42, `{42,42}`}},
			},
			{
				Statement: `select * from duplic('foo'::text);`,
				Results:   []sql.Row{{`foo`, `{foo,foo}`}},
			},
			{
				Statement: `drop function duplic(anyelement);`,
			},
			{
				Statement: `create function duplic(in i anycompatiblerange, out j anycompatible, out k anycompatiblearray) as $$
begin
  j := lower(i);`,
			},
			{
				Statement: `  k := array[lower(i),upper(i)];`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select * from duplic(int4range(42,49));`,
				Results:   []sql.Row{{42, `{42,49}`}},
			},
			{
				Statement: `select * from duplic(textrange('aaa', 'bbb'));`,
				Results:   []sql.Row{{`aaa`, `{aaa,bbb}`}},
			},
			{
				Statement: `drop function duplic(anycompatiblerange);`,
			},
			{
				Statement: `create table perform_test (
	a	INT,
	b	INT
);`,
			},
			{
				Statement: `create function perform_simple_func(int) returns boolean as '
BEGIN
	IF $1 < 20 THEN
		INSERT INTO perform_test VALUES ($1, $1 + 10);`,
			},
			{
				Statement: `		RETURN TRUE;`,
			},
			{
				Statement: `	ELSE
		RETURN FALSE;`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `create function perform_test_func() returns void as '
BEGIN
	IF FOUND then
		INSERT INTO perform_test VALUES (100, 100);`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `	PERFORM perform_simple_func(5);`,
			},
			{
				Statement: `	IF FOUND then
		INSERT INTO perform_test VALUES (100, 100);`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `	PERFORM perform_simple_func(50);`,
			},
			{
				Statement: `	IF FOUND then
		INSERT INTO perform_test VALUES (100, 100);`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `	RETURN;`,
			},
			{
				Statement: `END;' language plpgsql;`,
			},
			{
				Statement: `SELECT perform_test_func();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT * FROM perform_test;`,
				Results:   []sql.Row{{5, 15}, {100, 100}, {100, 100}},
			},
			{
				Statement: `drop table perform_test;`,
			},
			{
				Statement: `create temp table users(login text, id serial);`,
			},
			{
				Statement: `create function sp_id_user(a_login text) returns int as $$
declare x int;`,
			},
			{
				Statement: `begin
  select into x id from users where login = a_login;`,
			},
			{
				Statement: `  if found then return x; end if;`,
			},
			{
				Statement: `  return 0;`,
			},
			{
				Statement: `end$$ language plpgsql stable;`,
			},
			{
				Statement: `insert into users values('user1');`,
			},
			{
				Statement: `select sp_id_user('user1');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select sp_id_user('userx');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `create function sp_add_user(a_login text) returns int as $$
declare my_id_user int;`,
			},
			{
				Statement: `begin
  my_id_user = sp_id_user( a_login );`,
			},
			{
				Statement: `  IF  my_id_user > 0 THEN
    RETURN -1;  -- error code for existing user`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `  INSERT INTO users ( login ) VALUES ( a_login );`,
			},
			{
				Statement: `  my_id_user = sp_id_user( a_login );`,
			},
			{
				Statement: `  IF  my_id_user = 0 THEN
    RETURN -2;  -- error code for insertion failure`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `  RETURN my_id_user;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select sp_add_user('user1');`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `select sp_add_user('user2');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select sp_add_user('user2');`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `select sp_add_user('user3');`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select sp_add_user('user3');`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `drop function sp_add_user(text);`,
			},
			{
				Statement: `drop function sp_id_user(text);`,
			},
			{
				Statement: `create table rc_test (a int, b int);`,
			},
			{
				Statement: `copy rc_test from stdin;`,
			},
			{
				Statement: `create function return_unnamed_refcursor() returns refcursor as $$
declare
    rc refcursor;`,
			},
			{
				Statement: `begin
    open rc for select a from rc_test;`,
			},
			{
				Statement: `    return rc;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `create function use_refcursor(rc refcursor) returns int as $$
declare
    rc refcursor;`,
			},
			{
				Statement: `    x record;`,
			},
			{
				Statement: `begin
    rc := return_unnamed_refcursor();`,
			},
			{
				Statement: `    fetch next from rc into x;`,
			},
			{
				Statement: `    return x.a;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select use_refcursor(return_unnamed_refcursor());`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `create function return_refcursor(rc refcursor) returns refcursor as $$
begin
    open rc for select a from rc_test;`,
			},
			{
				Statement: `    return rc;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `create function refcursor_test1(refcursor) returns refcursor as $$
begin
    perform return_refcursor($1);`,
			},
			{
				Statement: `    return $1;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `select refcursor_test1('test1');`,
				Results:   []sql.Row{{`test1`}},
			},
			{
				Statement: `fetch next in test1;`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `select refcursor_test1('test2');`,
				Results:   []sql.Row{{`test2`}},
			},
			{
				Statement: `fetch all from test2;`,
				Results:   []sql.Row{{5}, {50}, {500}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement:   `fetch next from test1;`,
				ErrorString: `cursor "test1" does not exist`,
			},
			{
				Statement: `create function refcursor_test2(int, int) returns boolean as $$
declare
    c1 cursor (param1 int, param2 int) for select * from rc_test where a > param1 and b > param2;`,
			},
			{
				Statement: `    nonsense record;`,
			},
			{
				Statement: `begin
    open c1($1, $2);`,
			},
			{
				Statement: `    fetch c1 into nonsense;`,
			},
			{
				Statement: `    close c1;`,
			},
			{
				Statement: `    if found then
        return true;`,
			},
			{
				Statement: `    else
        return false;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select refcursor_test2(20000, 20000) as "Should be false",
       refcursor_test2(20, 20) as "Should be true";`,
				Results: []sql.Row{{false, true}},
			},
			{
				Statement: `create function constant_refcursor() returns refcursor as $$
declare
    rc constant refcursor;`,
			},
			{
				Statement: `begin
    open rc for select a from rc_test;`,
			},
			{
				Statement: `    return rc;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement:   `select constant_refcursor();`,
				ErrorString: `variable "rc" is declared CONSTANT`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function constant_refcursor() line 5 at OPEN
create or replace function constant_refcursor() returns refcursor as $$
declare
    rc constant refcursor := 'my_cursor_name';`,
			},
			{
				Statement: `begin
    open rc for select a from rc_test;`,
			},
			{
				Statement: `    return rc;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select constant_refcursor();`,
				Results:   []sql.Row{{`my_cursor_name`}},
			},
			{
				Statement: `create function namedparmcursor_test1(int, int) returns boolean as $$
declare
    c1 cursor (param1 int, param12 int) for select * from rc_test where a > param1 and b > param12;`,
			},
			{
				Statement: `    nonsense record;`,
			},
			{
				Statement: `begin
    open c1(param12 := $2, param1 := $1);`,
			},
			{
				Statement: `    fetch c1 into nonsense;`,
			},
			{
				Statement: `    close c1;`,
			},
			{
				Statement: `    if found then
        return true;`,
			},
			{
				Statement: `    else
        return false;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select namedparmcursor_test1(20000, 20000) as "Should be false",
       namedparmcursor_test1(20, 20) as "Should be true";`,
				Results: []sql.Row{{false, true}},
			},
			{
				Statement: `create function namedparmcursor_test2(int, int) returns boolean as $$
declare
    c1 cursor (param1 int, param2 int) for select * from rc_test where a > param1 and b > param2;`,
			},
			{
				Statement: `    nonsense record;`,
			},
			{
				Statement: `begin
    open c1(param1 := $1, $2);`,
			},
			{
				Statement: `    fetch c1 into nonsense;`,
			},
			{
				Statement: `    close c1;`,
			},
			{
				Statement: `    if found then
        return true;`,
			},
			{
				Statement: `    else
        return false;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select namedparmcursor_test2(20, 20);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create function namedparmcursor_test3() returns void as $$
declare
    c1 cursor (param1 int, param2 int) for select * from rc_test where a > param1 and b > param2;`,
			},
			{
				Statement: `begin
    open c1(param2 := 20, 21);`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
				ErrorString: `value for parameter "param2" of cursor "c1" specified more than once`,
			},
			{
				Statement: `create function namedparmcursor_test4() returns void as $$
declare
    c1 cursor (param1 int, param2 int) for select * from rc_test where a > param1 and b > param2;`,
			},
			{
				Statement: `begin
    open c1(20, param1 := 21);`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
				ErrorString: `value for parameter "param1" of cursor "c1" specified more than once`,
			},
			{
				Statement: `create function namedparmcursor_test5() returns void as $$
declare
  c1 cursor (p1 int, p2 int) for
    select * from tenk1 where thousand = p1 and tenthous = p2;`,
			},
			{
				Statement: `begin
  open c1 (p2 := 77, p2 := 42);`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
				ErrorString: `value for parameter "p2" of cursor "c1" specified more than once`,
			},
			{
				Statement: `create function namedparmcursor_test6() returns void as $$
declare
  c1 cursor (p1 int, p2 int) for
    select * from tenk1 where thousand = p1 and tenthous = p2;`,
			},
			{
				Statement: `begin
  open c1 (p2 := 77);`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
				ErrorString: `not enough arguments for cursor "c1"`,
			},
			{
				Statement: `create function namedparmcursor_test7() returns void as $$
declare
  c1 cursor (p1 int, p2 int) for
    select * from tenk1 where thousand = p1 and tenthous = p2;`,
			},
			{
				Statement: `begin
  open c1 (p2 := 77, p1 := 42/0);`,
			},
			{
				Statement: `end $$ language plpgsql;`,
			},
			{
				Statement:   `select namedparmcursor_test7();`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `CONTEXT:  SQL expression "42/0 AS p1, 77 AS p2"
PL/pgSQL function namedparmcursor_test7() line 6 at OPEN
create function namedparmcursor_test8() returns int4 as $$
declare
  c1 cursor (p1 int, p2 int) for
    select count(*) from tenk1 where thousand = p1 and tenthous = p2;`,
			},
			{
				Statement: `  n int4;`,
			},
			{
				Statement: `begin
  open c1 (77 -- test
  , 42);`,
			},
			{
				Statement: `  fetch c1 into n;`,
			},
			{
				Statement: `  return n;`,
			},
			{
				Statement: `end $$ language plpgsql;`,
			},
			{
				Statement: `select namedparmcursor_test8();`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `create function namedparmcursor_test9(p1 int) returns int4 as $$
declare
  c1 cursor (p1 int, p2 int, debug int) for
    select count(*) from tenk1 where thousand = p1 and tenthous = p2
      and four = debug;`,
			},
			{
				Statement: `  p2 int4 := 1006;`,
			},
			{
				Statement: `  n int4;`,
			},
			{
				Statement: `begin
  open c1 (p1 := p1, p2 := p2, debug := 2);`,
			},
			{
				Statement: `  fetch c1 into n;`,
			},
			{
				Statement: `  return n;`,
			},
			{
				Statement: `end $$ language plpgsql;`,
			},
			{
				Statement: `select namedparmcursor_test9(6);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `create function raise_test1(int) returns int as $$
begin
    raise notice 'This message has too many parameters!', $1;`,
			},
			{
				Statement: `    return $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$ language plpgsql;`,
				ErrorString: `too many parameters specified for RAISE`,
			},
			{
				Statement: `CONTEXT:  compilation of PL/pgSQL function "raise_test1" near line 3
create function raise_test2(int) returns int as $$
begin
    raise notice 'This message has too few parameters: %, %, %', $1, $1;`,
			},
			{
				Statement: `    return $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$ language plpgsql;`,
				ErrorString: `too few parameters specified for RAISE`,
			},
			{
				Statement: `CONTEXT:  compilation of PL/pgSQL function "raise_test2" near line 3
create function raise_test3(int) returns int as $$
begin
    raise notice 'This message has no parameters (despite having %% signs in it)!';`,
			},
			{
				Statement: `    return $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select raise_test3(1);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `CREATE FUNCTION reraise_test() RETURNS void AS $$
BEGIN
   BEGIN
       RAISE syntax_error;`,
			},
			{
				Statement: `   EXCEPTION
       WHEN syntax_error THEN
           BEGIN
               raise notice 'exception % thrown in inner block, reraising', sqlerrm;`,
			},
			{
				Statement: `               RAISE;`,
			},
			{
				Statement: `           EXCEPTION
               WHEN OTHERS THEN
                   raise notice 'RIGHT - exception % caught in inner block', sqlerrm;`,
			},
			{
				Statement: `           END;`,
			},
			{
				Statement: `   END;`,
			},
			{
				Statement: `EXCEPTION
   WHEN OTHERS THEN
       raise notice 'WRONG - exception % caught in outer block', sqlerrm;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT reraise_test();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create function bad_sql1() returns int as $$
declare a int;`,
			},
			{
				Statement: `begin
    a := 5;`,
			},
			{
				Statement: `    Johnny Yuma;`,
			},
			{
				Statement: `    a := 10;`,
			},
			{
				Statement: `    return a;`,
			},
			{
				Statement:   `end$$ language plpgsql;`,
				ErrorString: `syntax error at or near "Johnny"`,
			},
			{
				Statement: `create function bad_sql2() returns int as $$
declare r record;`,
			},
			{
				Statement: `begin
    for r in select I fought the law, the law won LOOP
        raise notice 'in loop';`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    return 5;`,
			},
			{
				Statement:   `end;$$ language plpgsql;`,
				ErrorString: `syntax error at or near "the"`,
			},
			{
				Statement: `create function missing_return_expr() returns int as $$
begin
    return ;`,
			},
			{
				Statement:   `end;$$ language plpgsql;`,
				ErrorString: `missing expression at or near ";"`,
			},
			{
				Statement: `create function void_return_expr() returns void as $$
begin
    return 5;`,
			},
			{
				Statement:   `end;$$ language plpgsql;`,
				ErrorString: `RETURN cannot have a parameter in function returning void`,
			},
			{
				Statement: `create function void_return_expr() returns void as $$
begin
    perform 2+2;`,
			},
			{
				Statement: `end;$$ language plpgsql;`,
			},
			{
				Statement: `select void_return_expr();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create function missing_return_expr() returns int as $$
begin
    perform 2+2;`,
			},
			{
				Statement: `end;$$ language plpgsql;`,
			},
			{
				Statement:   `select missing_return_expr();`,
				ErrorString: `control reached end of function without RETURN`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function missing_return_expr()
drop function void_return_expr();`,
			},
			{
				Statement: `drop function missing_return_expr();`,
			},
			{
				Statement: `create table eifoo (i integer, y integer);`,
			},
			{
				Statement: `create type eitype as (i integer, y integer);`,
			},
			{
				Statement: `create or replace function execute_into_test(varchar) returns record as $$
declare
    _r record;`,
			},
			{
				Statement: `    _rt eifoo%rowtype;`,
			},
			{
				Statement: `    _v eitype;`,
			},
			{
				Statement: `    i int;`,
			},
			{
				Statement: `    j int;`,
			},
			{
				Statement: `    k int;`,
			},
			{
				Statement: `begin
    execute 'insert into '||$1||' values(10,15)';`,
			},
			{
				Statement: `    execute 'select (row).* from (select row(10,1)::eifoo) s' into _r;`,
			},
			{
				Statement: `    raise notice '% %', _r.i, _r.y;`,
			},
			{
				Statement: `    execute 'select * from '||$1||' limit 1' into _rt;`,
			},
			{
				Statement: `    raise notice '% %', _rt.i, _rt.y;`,
			},
			{
				Statement: `    execute 'select *, 20 from '||$1||' limit 1' into i, j, k;`,
			},
			{
				Statement: `    raise notice '% % %', i, j, k;`,
			},
			{
				Statement: `    execute 'select 1,2' into _v;`,
			},
			{
				Statement: `    return _v;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement: `select execute_into_test('eifoo');`,
				Results:   []sql.Row{{`(1,2)`}},
			},
			{
				Statement: `drop table eifoo cascade;`,
			},
			{
				Statement: `drop type eitype cascade;`,
			},
			{
				Statement: `create function excpt_test1() returns void as $$
begin
    raise notice '% %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement:   `select excpt_test1();`,
				ErrorString: `column "sqlstate" does not exist`,
			},
			{
				Statement: `QUERY:  sqlstate
CONTEXT:  PL/pgSQL function excpt_test1() line 3 at RAISE
create function excpt_test2() returns void as $$
begin
    begin
        begin
            raise notice '% %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `        end;`,
			},
			{
				Statement: `    end;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement:   `select excpt_test2();`,
				ErrorString: `column "sqlstate" does not exist`,
			},
			{
				Statement: `QUERY:  sqlstate
CONTEXT:  PL/pgSQL function excpt_test2() line 5 at RAISE
create function excpt_test3() returns void as $$
begin
    begin
        raise exception 'user exception';`,
			},
			{
				Statement: `    exception when others then
	    raise notice 'caught exception % %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `	    begin
	        raise notice '% %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `	        perform 10/0;`,
			},
			{
				Statement: `        exception
            when substring_error then
                -- this exception handler shouldn't be invoked
                raise notice 'unexpected exception: % %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `	        when division_by_zero then
	            raise notice 'caught exception % %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `	    end;`,
			},
			{
				Statement: `	    raise notice '% %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `    end;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement: `select excpt_test3();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create function excpt_test4() returns text as $$
begin
	begin perform 1/0;`,
			},
			{
				Statement: `	exception when others then return sqlerrm; end;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement: `select excpt_test4();`,
				Results:   []sql.Row{{`division by zero`}},
			},
			{
				Statement: `drop function excpt_test1();`,
			},
			{
				Statement: `drop function excpt_test2();`,
			},
			{
				Statement: `drop function excpt_test3();`,
			},
			{
				Statement: `drop function excpt_test4();`,
			},
			{
				Statement: `create function raise_exprs() returns void as $$
declare
    a integer[] = '{10,20,30}';`,
			},
			{
				Statement: `    c varchar = 'xyz';`,
			},
			{
				Statement: `    i integer;`,
			},
			{
				Statement: `begin
    i := 2;`,
			},
			{
				Statement: `    raise notice '%; %; %; %; %; %', a, a[i], c, (select c || 'abc'), row(10,'aaa',NULL,30), NULL;`,
			},
			{
				Statement: `end;$$ language plpgsql;`,
			},
			{
				Statement: `select raise_exprs();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function raise_exprs();`,
			},
			{
				Statement: `create function multi_datum_use(p1 int) returns bool as $$
declare
  x int;`,
			},
			{
				Statement: `  y int;`,
			},
			{
				Statement: `begin
  select into x,y unique1/p1, unique1/$1 from tenk1 group by unique1/p1;`,
			},
			{
				Statement: `  return x = y;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select multi_datum_use(42);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create temp table foo (f1 int, f2 int);`,
			},
			{
				Statement: `insert into foo values (1,2), (3,4);`,
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should work
  insert into foo values(5,6) returning * into x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select stricttest();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should fail due to implicit strict
  insert into foo values(7,8),(9,10) returning * into x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at SQL statement
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should work
  execute 'insert into foo values(5,6) returning *' into x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select stricttest();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- this should work since EXECUTE isn't as picky
  execute 'insert into foo values(7,8),(9,10) returning *' into x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select stricttest();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from foo;`,
				Results:   []sql.Row{{1, 2}, {3, 4}, {5, 6}, {5, 6}, {7, 8}, {9, 10}},
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should work
  select * from foo where f1 = 3 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select stricttest();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should fail, no rows
  select * from foo where f1 = 0 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned no rows`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at SQL statement
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should fail, too many rows
  select * from foo where f1 > 3 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at SQL statement
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should work
  execute 'select * from foo where f1 = 3' into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select stricttest();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should fail, no rows
  execute 'select * from foo where f1 = 0' into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned no rows`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at EXECUTE
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- should fail, too many rows
  execute 'select * from foo where f1 > 3' into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at EXECUTE
drop function stricttest();`,
			},
			{
				Statement: `set plpgsql.print_strict_params to true;`,
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
declare
x record;`,
			},
			{
				Statement: `p1 int := 2;`,
			},
			{
				Statement: `p3 text := 'foo';`,
			},
			{
				Statement: `begin
  -- no rows
  select * from foo where f1 = p1 and f1::text = p3 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned no rows`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 8 at SQL statement
create or replace function stricttest() returns void as $$
declare
x record;`,
			},
			{
				Statement: `p1 int := 2;`,
			},
			{
				Statement: `p3 text := $a$'Valame Dios!' dijo Sancho; 'no le dije yo a vuestra merced que mirase bien lo que hacia?'$a$;`,
			},
			{
				Statement: `begin
  -- no rows
  select * from foo where f1 = p1 and f1::text = p3 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned no rows`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 8 at SQL statement
create or replace function stricttest() returns void as $$
declare
x record;`,
			},
			{
				Statement: `p1 int := 2;`,
			},
			{
				Statement: `p3 text := 'foo';`,
			},
			{
				Statement: `begin
  -- too many rows
  select * from foo where f1 > p1 or f1::text = p3  into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 8 at SQL statement
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- too many rows, no params
  select * from foo where f1 > 3 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at SQL statement
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- no rows
  execute 'select * from foo where f1 = $1 or f1::text = $2' using 0, 'foo' into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned no rows`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at EXECUTE
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- too many rows
  execute 'select * from foo where f1 > $1' using 1 into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at EXECUTE
create or replace function stricttest() returns void as $$
declare x record;`,
			},
			{
				Statement: `begin
  -- too many rows, no parameters
  execute 'select * from foo where f1 > 3' into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 5 at EXECUTE
create or replace function stricttest() returns void as $$
#print_strict_params off
declare
x record;`,
			},
			{
				Statement: `p1 int := 2;`,
			},
			{
				Statement: `p3 text := 'foo';`,
			},
			{
				Statement: `begin
  -- too many rows
  select * from foo where f1 > p1 or f1::text = p3  into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 10 at SQL statement
reset plpgsql.print_strict_params;`,
			},
			{
				Statement: `create or replace function stricttest() returns void as $$
#print_strict_params on
declare
x record;`,
			},
			{
				Statement: `p1 int := 2;`,
			},
			{
				Statement: `p3 text := 'foo';`,
			},
			{
				Statement: `begin
  -- too many rows
  select * from foo where f1 > p1 or f1::text = p3  into strict x;`,
			},
			{
				Statement: `  raise notice 'x.f1 = %, x.f2 = %', x.f1, x.f2;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select stricttest();`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stricttest() line 10 at SQL statement
set plpgsql.extra_warnings to 'all';`,
			},
			{
				Statement: `set plpgsql.extra_warnings to 'none';`,
			},
			{
				Statement: `set plpgsql.extra_errors to 'all';`,
			},
			{
				Statement: `set plpgsql.extra_errors to 'none';`,
			},
			{
				Statement: `set plpgsql.extra_warnings to 'shadowed_variables';`,
			},
			{
				Statement: `create or replace function shadowtest(in1 int)
	returns table (out1 int) as $$
declare
in1 int;`,
			},
			{
				Statement: `out1 int;`,
			},
			{
				Statement: `begin
end
$$ language plpgsql;`,
			},
			{
				Statement: `LINE 4: in1 int;`,
			},
			{
				Statement: `        ^
LINE 5: out1 int;`,
			},
			{
				Statement: `        ^
select shadowtest(1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `set plpgsql.extra_warnings to 'shadowed_variables';`,
			},
			{
				Statement: `select shadowtest(1);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `create or replace function shadowtest(in1 int)
	returns table (out1 int) as $$
declare
in1 int;`,
			},
			{
				Statement: `out1 int;`,
			},
			{
				Statement: `begin
end
$$ language plpgsql;`,
			},
			{
				Statement: `LINE 4: in1 int;`,
			},
			{
				Statement: `        ^
LINE 5: out1 int;`,
			},
			{
				Statement: `        ^
select shadowtest(1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `drop function shadowtest(int);`,
			},
			{
				Statement: `create or replace function shadowtest()
	returns void as $$
declare
f1 int;`,
			},
			{
				Statement: `begin
	declare
	f1 int;`,
			},
			{
				Statement: `	begin
	end;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `LINE 7:  f1 int;`,
			},
			{
				Statement: `         ^
drop function shadowtest();`,
			},
			{
				Statement: `create or replace function shadowtest(in1 int)
	returns void as $$
declare
in1 int;`,
			},
			{
				Statement: `begin
	declare
	in1 int;`,
			},
			{
				Statement: `	begin
	end;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `LINE 4: in1 int;`,
			},
			{
				Statement: `        ^
LINE 7:  in1 int;`,
			},
			{
				Statement: `         ^
drop function shadowtest(int);`,
			},
			{
				Statement: `create or replace function shadowtest()
	returns void as $$
declare
f1 int;`,
			},
			{
				Statement: `c1 cursor (f1 int) for select 1;`,
			},
			{
				Statement: `begin
end$$ language plpgsql;`,
			},
			{
				Statement: `LINE 5: c1 cursor (f1 int) for select 1;`,
			},
			{
				Statement: `                   ^
drop function shadowtest();`,
			},
			{
				Statement: `set plpgsql.extra_errors to 'shadowed_variables';`,
			},
			{
				Statement: `create or replace function shadowtest(f1 int)
	returns boolean as $$
declare f1 int; begin return 1; end $$ language plpgsql;`,
				ErrorString: `variable "f1" shadows a previously defined variable`,
			},
			{
				Statement:   `select shadowtest(1);`,
				ErrorString: `function shadowtest(integer) does not exist`,
			},
			{
				Statement: `reset plpgsql.extra_errors;`,
			},
			{
				Statement: `reset plpgsql.extra_warnings;`,
			},
			{
				Statement: `create or replace function shadowtest(f1 int)
	returns boolean as $$
declare f1 int; begin return 1; end $$ language plpgsql;`,
			},
			{
				Statement: `select shadowtest(1);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `set plpgsql.extra_warnings to 'too_many_rows';`,
			},
			{
				Statement: `do $$
declare x int;`,
			},
			{
				Statement: `begin
  select v from generate_series(1,2) g(v) into x;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `HINT:  Make sure the query returns a single row, or use LIMIT 1.
set plpgsql.extra_errors to 'too_many_rows';`,
			},
			{
				Statement: `do $$
declare x int;`,
			},
			{
				Statement: `begin
  select v from generate_series(1,2) g(v) into x;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `query returned more than one row`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 4 at SQL statement
reset plpgsql.extra_errors;`,
			},
			{
				Statement: `reset plpgsql.extra_warnings;`,
			},
			{
				Statement: `set plpgsql.extra_warnings to 'strict_multi_assignment';`,
			},
			{
				Statement: `do $$
declare
  x int;`,
			},
			{
				Statement: `  y int;`,
			},
			{
				Statement: `begin
  select 1 into x, y;`,
			},
			{
				Statement: `  select 1,2 into x, y;`,
			},
			{
				Statement: `  select 1,2,3 into x, y;`,
			},
			{
				Statement: `end
$$;`,
			},
			{
				Statement: `HINT:  Make sure the query returns the exact list of columns.
HINT:  Make sure the query returns the exact list of columns.
set plpgsql.extra_errors to 'strict_multi_assignment';`,
			},
			{
				Statement: `do $$
declare
  x int;`,
			},
			{
				Statement: `  y int;`,
			},
			{
				Statement: `begin
  select 1 into x, y;`,
			},
			{
				Statement: `  select 1,2 into x, y;`,
			},
			{
				Statement: `  select 1,2,3 into x, y;`,
			},
			{
				Statement: `end
$$;`,
				ErrorString: `number of source and target fields in assignment does not match`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 6 at SQL statement
create table test_01(a int, b int, c int);`,
			},
			{
				Statement: `alter table test_01 drop column a;`,
			},
			{
				Statement: `insert into test_01 values(10,20);`,
			},
			{
				Statement: `do $$
declare
  x int;`,
			},
			{
				Statement: `  y int;`,
			},
			{
				Statement: `begin
  select * from test_01 into x, y; -- should be ok`,
			},
			{
				Statement: `  raise notice 'ok';`,
			},
			{
				Statement: `  select * from test_01 into x;    -- should to fail`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `number of source and target fields in assignment does not match`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 8 at SQL statement
do $$
declare
  t test_01;`,
			},
			{
				Statement: `begin
  select 1, 2 into t;  -- should be ok`,
			},
			{
				Statement: `  raise notice 'ok';`,
			},
			{
				Statement: `  select 1, 2, 3 into t; -- should fail;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `number of source and target fields in assignment does not match`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 7 at SQL statement
do $$
declare
  t test_01;`,
			},
			{
				Statement: `begin
  select 1 into t; -- should fail;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `number of source and target fields in assignment does not match`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 5 at SQL statement
drop table test_01;`,
			},
			{
				Statement: `reset plpgsql.extra_errors;`,
			},
			{
				Statement: `reset plpgsql.extra_warnings;`,
			},
			{
				Statement: `create function sc_test() returns setof integer as $$
declare
  c scroll cursor for select f1 from int4_tbl;`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c;`,
			},
			{
				Statement: `  fetch last from c into x;`,
			},
			{
				Statement: `  while found loop
    return next x;`,
			},
			{
				Statement: `    fetch prior from c into x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from sc_test();`,
				Results:   []sql.Row{{-2147483647}, {2147483647}, {-123456}, {123456}, {0}},
			},
			{
				Statement: `create or replace function sc_test() returns setof integer as $$
declare
  c no scroll cursor for select f1 from int4_tbl;`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c;`,
			},
			{
				Statement: `  fetch last from c into x;`,
			},
			{
				Statement: `  while found loop
    return next x;`,
			},
			{
				Statement: `    fetch prior from c into x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select * from sc_test();  -- fails because of NO SCROLL specification`,
				ErrorString: `cursor can only scan forward`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function sc_test() line 7 at FETCH
create or replace function sc_test() returns setof integer as $$
declare
  c refcursor;`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c scroll for select f1 from int4_tbl;`,
			},
			{
				Statement: `  fetch last from c into x;`,
			},
			{
				Statement: `  while found loop
    return next x;`,
			},
			{
				Statement: `    fetch prior from c into x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from sc_test();`,
				Results:   []sql.Row{{-2147483647}, {2147483647}, {-123456}, {123456}, {0}},
			},
			{
				Statement: `create or replace function sc_test() returns setof integer as $$
declare
  c refcursor;`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c scroll for execute 'select f1 from int4_tbl';`,
			},
			{
				Statement: `  fetch last from c into x;`,
			},
			{
				Statement: `  while found loop
    return next x;`,
			},
			{
				Statement: `    fetch relative -2 from c into x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from sc_test();`,
				Results:   []sql.Row{{-2147483647}, {-123456}, {0}},
			},
			{
				Statement: `create or replace function sc_test() returns setof integer as $$
declare
  c refcursor;`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c scroll for execute 'select f1 from int4_tbl';`,
			},
			{
				Statement: `  fetch last from c into x;`,
			},
			{
				Statement: `  while found loop
    return next x;`,
			},
			{
				Statement: `    move backward 2 from c;`,
			},
			{
				Statement: `    fetch relative -1 from c into x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from sc_test();`,
				Results:   []sql.Row{{-2147483647}, {123456}},
			},
			{
				Statement: `create or replace function sc_test() returns setof integer as $$
declare
  c cursor for select * from generate_series(1, 10);`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c;`,
			},
			{
				Statement: `  loop
      move relative 2 in c;`,
			},
			{
				Statement: `      if not found then
          exit;`,
			},
			{
				Statement: `      end if;`,
			},
			{
				Statement: `      fetch next from c into x;`,
			},
			{
				Statement: `      if found then
          return next x;`,
			},
			{
				Statement: `      end if;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from sc_test();`,
				Results:   []sql.Row{{3}, {6}, {9}},
			},
			{
				Statement: `create or replace function sc_test() returns setof integer as $$
declare
  c cursor for select * from generate_series(1, 10);`,
			},
			{
				Statement: `  x integer;`,
			},
			{
				Statement: `begin
  open c;`,
			},
			{
				Statement: `  move forward all in c;`,
			},
			{
				Statement: `  fetch backward from c into x;`,
			},
			{
				Statement: `  if found then
    return next x;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from sc_test();`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `drop function sc_test();`,
			},
			{
				Statement: `create function pl_qual_names (param1 int) returns void as $$
<<outerblock>>
declare
  param1 int := 1;`,
			},
			{
				Statement: `begin
  <<innerblock>>
  declare
    param1 int := 2;`,
			},
			{
				Statement: `  begin
    raise notice 'param1 = %', param1;`,
			},
			{
				Statement: `    raise notice 'pl_qual_names.param1 = %', pl_qual_names.param1;`,
			},
			{
				Statement: `    raise notice 'outerblock.param1 = %', outerblock.param1;`,
			},
			{
				Statement: `    raise notice 'innerblock.param1 = %', innerblock.param1;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select pl_qual_names(42);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function pl_qual_names(int);`,
			},
			{
				Statement: `create function ret_query1(out int, out int) returns setof record as $$
begin
    $1 := -1;`,
			},
			{
				Statement: `    $2 := -2;`,
			},
			{
				Statement: `    return next;`,
			},
			{
				Statement: `    return query select x + 1, x * 10 from generate_series(0, 10) s (x);`,
			},
			{
				Statement: `    return next;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from ret_query1();`,
				Results:   []sql.Row{{-1, -2}, {1, 0}, {2, 10}, {3, 20}, {4, 30}, {5, 40}, {6, 50}, {7, 60}, {8, 70}, {9, 80}, {10, 90}, {11, 100}, {-1, -2}},
			},
			{
				Statement: `create type record_type as (x text, y int, z boolean);`,
			},
			{
				Statement: `create or replace function ret_query2(lim int) returns setof record_type as $$
begin
    return query select md5(s.x::text), s.x, s.x > 0
                 from generate_series(-8, lim) s (x) where s.x % 2 = 0;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from ret_query2(8);`,
				Results:   []sql.Row{{`a8d2ec85eaf98407310b72eb73dda247`, -8, false}, {`596a3d04481816330f07e4f97510c28f`, -6, false}, {`0267aaf632e87a63288a08331f22c7c3`, -4, false}, {`5d7b9adcbe1c629ec722529dd12e5129`, -2, false}, {`cfcd208495d565ef66e7dff9f98764da`, 0, false}, {`c81e728d9d4c2f636f067f89cc14862c`, 2, true}, {`a87ff679a2f3e71d9181a67b7542122c`, 4, true}, {`1679091c5a880faf6fb5e6087eb1b2dc`, 6, true}, {`c9f0f895fb98ab9159f51fd0297e236d`, 8, true}},
			},
			{
				Statement: `create function exc_using(int, text) returns int as $$
declare i int;`,
			},
			{
				Statement: `begin
  for i in execute 'select * from generate_series(1,$1)' using $1+1 loop
    raise notice '%', i;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  execute 'select $2 + $2*3 + length($1)' into i using $2,$1;`,
			},
			{
				Statement: `  return i;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select exc_using(5, 'foobar');`,
				Results:   []sql.Row{{26}},
			},
			{
				Statement: `drop function exc_using(int, text);`,
			},
			{
				Statement: `create or replace function exc_using(int) returns void as $$
declare
  c refcursor;`,
			},
			{
				Statement: `  i int;`,
			},
			{
				Statement: `begin
  open c for execute 'select * from generate_series(1,$1)' using $1+1;`,
			},
			{
				Statement: `  loop
    fetch c into i;`,
			},
			{
				Statement: `    exit when not found;`,
			},
			{
				Statement: `    raise notice '%', i;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  close c;`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select exc_using(5);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function exc_using(int);`,
			},
			{
				Statement: `create or replace function forc01() returns void as $$
declare
  c cursor(r1 integer, r2 integer)
       for select * from generate_series(r1,r2) i;`,
			},
			{
				Statement: `  c2 cursor
       for select * from generate_series(41,43) i;`,
			},
			{
				Statement: `begin
  for r in c(5,7) loop
    raise notice '% from %', r.i, c;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  -- again, to test if cursor was closed properly
  for r in c(9,10) loop
    raise notice '% from %', r.i, c;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  -- and test a parameterless cursor
  for r in c2 loop
    raise notice '% from %', r.i, c2;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  -- and try it with a hand-assigned name
  raise notice 'after loop, c2 = %', c2;`,
			},
			{
				Statement: `  c2 := 'special_name';`,
			},
			{
				Statement: `  for r in c2 loop
    raise notice '% from %', r.i, c2;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  raise notice 'after loop, c2 = %', c2;`,
			},
			{
				Statement: `  -- and try it with a generated name
  -- (which we can't show in the output because it's variable)
  c2 := null;`,
			},
			{
				Statement: `  for r in c2 loop
    raise notice '%', r.i;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  raise notice 'after loop, c2 = %', c2;`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select forc01();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create temp table forc_test as
  select n as i, n as j from generate_series(1,10) n;`,
			},
			{
				Statement: `create or replace function forc01() returns void as $$
declare
  c cursor for select * from forc_test;`,
			},
			{
				Statement: `begin
  for r in c loop
    raise notice '%, %', r.i, r.j;`,
			},
			{
				Statement: `    update forc_test set i = i * 100, j = r.j * 2 where current of c;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select forc01();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from forc_test;`,
				Results:   []sql.Row{{100, 2}, {200, 4}, {300, 6}, {400, 8}, {500, 10}, {600, 12}, {700, 14}, {800, 16}, {900, 18}, {1000, 20}},
			},
			{
				Statement: `create or replace function forc01() returns void as $$
declare
  c refcursor := 'fooled_ya';`,
			},
			{
				Statement: `  r record;`,
			},
			{
				Statement: `begin
  open c for select * from forc_test;`,
			},
			{
				Statement: `  loop
    fetch c into r;`,
			},
			{
				Statement: `    exit when not found;`,
			},
			{
				Statement: `    raise notice '%, %', r.i, r.j;`,
			},
			{
				Statement: `    update forc_test set i = i * 100, j = r.j * 2 where current of c;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select forc01();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from forc_test;`,
				Results:   []sql.Row{{10000, 4}, {20000, 8}, {30000, 12}, {40000, 16}, {50000, 20}, {60000, 24}, {70000, 28}, {80000, 32}, {90000, 36}, {100000, 40}},
			},
			{
				Statement: `drop function forc01();`,
			},
			{
				Statement: `create or replace function forc_bad() returns void as $$
declare
  c refcursor;`,
			},
			{
				Statement: `begin
  for r in c loop
    raise notice '%', r.i;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$ language plpgsql;`,
				ErrorString: `cursor FOR loop must use a bound cursor variable`,
			},
			{
				Statement: `create or replace function return_dquery()
returns setof int as $$
begin
  return query execute 'select * from (values(10),(20)) f';`,
			},
			{
				Statement: `  return query execute 'select * from (values($1),($2)) f' using 40,50;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from return_dquery();`,
				Results:   []sql.Row{{10}, {20}, {40}, {50}},
			},
			{
				Statement: `drop function return_dquery();`,
			},
			{
				Statement: `create table tabwithcols(a int, b int, c int, d int);`,
			},
			{
				Statement: `insert into tabwithcols values(10,20,30,40),(50,60,70,80);`,
			},
			{
				Statement: `create or replace function returnqueryf()
returns setof tabwithcols as $$
begin
  return query select * from tabwithcols;`,
			},
			{
				Statement: `  return query execute 'select * from tabwithcols';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from returnqueryf();`,
				Results:   []sql.Row{{10, 20, 30, 40}, {50, 60, 70, 80}, {10, 20, 30, 40}, {50, 60, 70, 80}},
			},
			{
				Statement: `alter table tabwithcols drop column b;`,
			},
			{
				Statement: `select * from returnqueryf();`,
				Results:   []sql.Row{{10, 30, 40}, {50, 70, 80}, {10, 30, 40}, {50, 70, 80}},
			},
			{
				Statement: `alter table tabwithcols drop column d;`,
			},
			{
				Statement: `select * from returnqueryf();`,
				Results:   []sql.Row{{10, 30}, {50, 70}, {10, 30}, {50, 70}},
			},
			{
				Statement: `alter table tabwithcols add column d int;`,
			},
			{
				Statement: `select * from returnqueryf();`,
				Results:   []sql.Row{{10, 30, ``}, {50, 70, ``}, {10, 30, ``}, {50, 70, ``}},
			},
			{
				Statement: `drop function returnqueryf();`,
			},
			{
				Statement: `drop table tabwithcols;`,
			},
			{
				Statement: `create type compostype as (x int, y varchar);`,
			},
			{
				Statement: `create or replace function compos() returns compostype as $$
declare
  v compostype;`,
			},
			{
				Statement: `begin
  v := (1, 'hello');`,
			},
			{
				Statement: `  return v;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select compos();`,
				Results:   []sql.Row{{`(1,hello)`}},
			},
			{
				Statement: `create or replace function compos() returns compostype as $$
declare
  v record;`,
			},
			{
				Statement: `begin
  v := (1, 'hello'::varchar);`,
			},
			{
				Statement: `  return v;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select compos();`,
				Results:   []sql.Row{{`(1,hello)`}},
			},
			{
				Statement: `create or replace function compos() returns compostype as $$
begin
  return (1, 'hello'::varchar);`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select compos();`,
				Results:   []sql.Row{{`(1,hello)`}},
			},
			{
				Statement: `create or replace function compos() returns compostype as $$
begin
  return (1, 'hello');`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select compos();`,
				ErrorString: `returned record type does not match expected record type`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function compos() while casting return value to function's return type
create or replace function compos() returns compostype as $$
begin
  return (1, 'hello')::compostype;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select compos();`,
				Results:   []sql.Row{{`(1,hello)`}},
			},
			{
				Statement: `drop function compos();`,
			},
			{
				Statement: `create or replace function composrec() returns record as $$
declare
  v record;`,
			},
			{
				Statement: `begin
  v := (1, 'hello');`,
			},
			{
				Statement: `  return v;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select composrec();`,
				Results:   []sql.Row{{`(1,hello)`}},
			},
			{
				Statement: `create or replace function composrec() returns record as $$
begin
  return (1, 'hello');`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select composrec();`,
				Results:   []sql.Row{{`(1,hello)`}},
			},
			{
				Statement: `drop function composrec();`,
			},
			{
				Statement: `create or replace function compos() returns setof compostype as $$
begin
  for i in 1..3
  loop
    return next (1, 'hello'::varchar);`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  return next null::compostype;`,
			},
			{
				Statement: `  return next (2, 'goodbye')::compostype;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from compos();`,
				Results:   []sql.Row{{1, `hello`}, {1, `hello`}, {1, `hello`}, {``, ``}, {2, `goodbye`}},
			},
			{
				Statement: `drop function compos();`,
			},
			{
				Statement: `create or replace function compos() returns compostype as $$
begin
  return 1 + 1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select compos();`,
				ErrorString: `cannot return non-composite value from function returning composite type`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function compos() line 3 at RETURN
create or replace function compos() returns compostype as $$
declare x int := 42;`,
			},
			{
				Statement: `begin
  return x;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select * from compos();`,
				ErrorString: `cannot return non-composite value from function returning composite type`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function compos() line 4 at RETURN
drop function compos();`,
			},
			{
				Statement: `create or replace function compos() returns int as $$
declare
  v compostype;`,
			},
			{
				Statement: `begin
  v := (1, 'hello');`,
			},
			{
				Statement: `  return v;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select compos();`,
				ErrorString: `invalid input syntax for type integer: "(1,hello)"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function compos() while casting return value to function's return type
create or replace function compos() returns int as $$
begin
  return (1, 'hello')::compostype;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select compos();`,
				ErrorString: `invalid input syntax for type integer: "(1,hello)"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function compos() while casting return value to function's return type
drop function compos();`,
			},
			{
				Statement: `drop type compostype;`,
			},
			{
				Statement: `create or replace function raise_test() returns void as $$
begin
  raise notice '% % %', 1, 2, 3
     using errcode = '55001', detail = 'some detail info', hint = 'some hint';`,
			},
			{
				Statement: `  raise '% % %', 1, 2, 3
     using errcode = 'division_by_zero', detail = 'some detail info';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select raise_test();`,
			},
			{
				Statement: `HINT:  some hint
ERROR:  1 2 3
CONTEXT:  PL/pgSQL function raise_test() line 5 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise 'check me'
     using errcode = 'division_by_zero', detail = 'some detail info';`,
			},
			{
				Statement: `  exception
    when others then
      raise notice 'SQLSTATE: % SQLERRM: %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `      raise;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `check me`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise 'check me'
     using errcode = '1234F', detail = 'some detail info';`,
			},
			{
				Statement: `  exception
    when others then
      raise notice 'SQLSTATE: % SQLERRM: %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `      raise;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `check me`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise 'check me'
     using errcode = '1234F', detail = 'some detail info';`,
			},
			{
				Statement: `  exception
    when sqlstate '1234F' then
      raise notice 'SQLSTATE: % SQLERRM: %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `      raise;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `check me`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise division_by_zero using detail = 'some detail info';`,
			},
			{
				Statement: `  exception
    when others then
      raise notice 'SQLSTATE: % SQLERRM: %', sqlstate, sqlerrm;`,
			},
			{
				Statement: `      raise;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `division_by_zero`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise division_by_zero;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `division_by_zero`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise sqlstate '1234F';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `1234F`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise division_by_zero using message = 'custom' || ' message';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `custom message`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise using message = 'custom' || ' message', errcode = '22012';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `custom message`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise notice 'some message' using message = 'custom' || ' message', errcode = '22012';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `RAISE option already specified: MESSAGE`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise division_by_zero using message = 'custom' || ' message', errcode = '22012';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `RAISE option already specified: ERRCODE`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create or replace function raise_test() returns void as $$
begin
  raise;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `RAISE without parameters cannot be used outside an exception handler`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 3 at RAISE
create function zero_divide() returns int as $$
declare v int := 0;`,
			},
			{
				Statement: `begin
  return 10 / v;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create or replace function raise_test() returns void as $$
begin
  raise exception 'custom exception'
     using detail = 'some detail of custom exception',
           hint = 'some hint related to custom exception';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create function stacked_diagnostics_test() returns void as $$
declare _sqlstate text;`,
			},
			{
				Statement: `        _message text;`,
			},
			{
				Statement: `        _context text;`,
			},
			{
				Statement: `begin
  perform zero_divide();`,
			},
			{
				Statement: `exception when others then
  get stacked diagnostics
        _sqlstate = returned_sqlstate,
        _message = message_text,
        _context = pg_exception_context;`,
			},
			{
				Statement: `  raise notice 'sqlstate: %, message: %, context: [%]',
    _sqlstate, _message, replace(_context, E'\n', ' <- ');`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select stacked_diagnostics_test();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function stacked_diagnostics_test() returns void as $$
declare _detail text;`,
			},
			{
				Statement: `        _hint text;`,
			},
			{
				Statement: `        _message text;`,
			},
			{
				Statement: `begin
  perform raise_test();`,
			},
			{
				Statement: `exception when others then
  get stacked diagnostics
        _message = message_text,
        _detail = pg_exception_detail,
        _hint = pg_exception_hint;`,
			},
			{
				Statement: `  raise notice 'message: %, detail: %, hint: %', _message, _detail, _hint;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select stacked_diagnostics_test();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function stacked_diagnostics_test() returns void as $$
declare _detail text;`,
			},
			{
				Statement: `        _hint text;`,
			},
			{
				Statement: `        _message text;`,
			},
			{
				Statement: `begin
  get stacked diagnostics
        _message = message_text,
        _detail = pg_exception_detail,
        _hint = pg_exception_hint;`,
			},
			{
				Statement: `  raise notice 'message: %, detail: %, hint: %', _message, _detail, _hint;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select stacked_diagnostics_test();`,
				ErrorString: `GET STACKED DIAGNOSTICS cannot be used outside an exception handler`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function stacked_diagnostics_test() line 6 at GET STACKED DIAGNOSTICS
drop function zero_divide();`,
			},
			{
				Statement: `drop function stacked_diagnostics_test();`,
			},
			{
				Statement: `create or replace function raise_test() returns void as $$
begin
  perform 1/0;`,
			},
			{
				Statement: `exception
  when sqlstate '22012' then
    raise notice using message = sqlstate;`,
			},
			{
				Statement: `    raise sqlstate '22012' using message = 'substitute message';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select raise_test();`,
				ErrorString: `substitute message`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function raise_test() line 7 at RAISE
drop function raise_test();`,
			},
			{
				Statement: `create or replace function stacked_diagnostics_test() returns void as $$
declare _column_name text;`,
			},
			{
				Statement: `        _constraint_name text;`,
			},
			{
				Statement: `        _datatype_name text;`,
			},
			{
				Statement: `        _table_name text;`,
			},
			{
				Statement: `        _schema_name text;`,
			},
			{
				Statement: `begin
  raise exception using
    column = '>>some column name<<',
    constraint = '>>some constraint name<<',
    datatype = '>>some datatype name<<',
    table = '>>some table name<<',
    schema = '>>some schema name<<';`,
			},
			{
				Statement: `exception when others then
  get stacked diagnostics
        _column_name = column_name,
        _constraint_name = constraint_name,
        _datatype_name = pg_datatype_name,
        _table_name = table_name,
        _schema_name = schema_name;`,
			},
			{
				Statement: `  raise notice 'column %, constraint %, type %, table %, schema %',
    _column_name, _constraint_name, _datatype_name, _table_name, _schema_name;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select stacked_diagnostics_test();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function stacked_diagnostics_test();`,
			},
			{
				Statement: `create or replace function vari(variadic int[])
returns void as $$
begin
  for i in array_lower($1,1)..array_upper($1,1) loop
    raise notice '%', $1[i];`,
			},
			{
				Statement: `  end loop; end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select vari(1,2,3,4,5);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select vari(3,4,5);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select vari(variadic array[5,6,7]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function vari(int[]);`,
			},
			{
				Statement: `create or replace function pleast(variadic numeric[])
returns numeric as $$
declare aux numeric = $1[array_lower($1,1)];`,
			},
			{
				Statement: `begin
  for i in array_lower($1,1)+1..array_upper($1,1) loop
    if $1[i] < aux then aux := $1[i]; end if;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  return aux;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql immutable strict;`,
			},
			{
				Statement: `select pleast(10,1,2,3,-16);`,
				Results:   []sql.Row{{-16}},
			},
			{
				Statement: `select pleast(10.2,2.2,-1.1);`,
				Results:   []sql.Row{{-1.1}},
			},
			{
				Statement: `select pleast(10.2,10, -20);`,
				Results:   []sql.Row{{-20}},
			},
			{
				Statement: `select pleast(10,20, -1.0);`,
				Results:   []sql.Row{{-1.0}},
			},
			{
				Statement: `create or replace function pleast(numeric)
returns numeric as $$
begin
  raise notice 'non-variadic function called';`,
			},
			{
				Statement: `  return $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql immutable strict;`,
			},
			{
				Statement: `select pleast(10);`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `drop function pleast(numeric[]);`,
			},
			{
				Statement: `drop function pleast(numeric);`,
			},
			{
				Statement: `create function tftest(int) returns table(a int, b int) as $$
begin
  return query select $1, $1+i from generate_series(1,5) g(i);`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql immutable strict;`,
			},
			{
				Statement: `select * from tftest(10);`,
				Results:   []sql.Row{{10, 11}, {10, 12}, {10, 13}, {10, 14}, {10, 15}},
			},
			{
				Statement: `create or replace function tftest(a1 int) returns table(a int, b int) as $$
begin
  a := a1; b := a1 + 1;`,
			},
			{
				Statement: `  return next;`,
			},
			{
				Statement: `  a := a1 * 10; b := a1 * 10 + 1;`,
			},
			{
				Statement: `  return next;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql immutable strict;`,
			},
			{
				Statement: `select * from tftest(10);`,
				Results:   []sql.Row{{10, 11}, {100, 101}},
			},
			{
				Statement: `drop function tftest(int);`,
			},
			{
				Statement: `create function rttest()
returns setof int as $$
declare rc int;`,
			},
			{
				Statement: `begin
  return query values(10),(20);`,
			},
			{
				Statement: `  get diagnostics rc = row_count;`,
			},
			{
				Statement: `  raise notice '% %', found, rc;`,
			},
			{
				Statement: `  return query select * from (values(10),(20)) f(a) where false;`,
			},
			{
				Statement: `  get diagnostics rc = row_count;`,
			},
			{
				Statement: `  raise notice '% %', found, rc;`,
			},
			{
				Statement: `  return query execute 'values(10),(20)';`,
			},
			{
				Statement: `  get diagnostics rc = row_count;`,
			},
			{
				Statement: `  raise notice '% %', found, rc;`,
			},
			{
				Statement: `  return query execute 'select * from (values(10),(20)) f(a) where false';`,
			},
			{
				Statement: `  get diagnostics rc = row_count;`,
			},
			{
				Statement: `  raise notice '% %', found, rc;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from rttest();`,
				Results:   []sql.Row{{10}, {20}, {10}, {20}},
			},
			{
				Statement: `create or replace function rttest()
returns setof int as $$
begin
  return query select 10 into no_such_table;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select * from rttest();`,
				ErrorString: `SELECT INTO query does not return tuples`,
			},
			{
				Statement: `CONTEXT:  SQL statement "select 10 into no_such_table"
PL/pgSQL function rttest() line 3 at RETURN QUERY
create or replace function rttest()
returns setof int as $$
begin
  return query execute 'select 10 into no_such_table';`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select * from rttest();`,
				ErrorString: `SELECT INTO query does not return tuples`,
			},
			{
				Statement: `CONTEXT:  SQL statement "select 10 into no_such_table"
PL/pgSQL function rttest() line 3 at RETURN QUERY
select * from no_such_table;`,
				ErrorString: `relation "no_such_table" does not exist`,
			},
			{
				Statement: `drop function rttest();`,
			},
			{
				Statement: `CREATE FUNCTION leaker_1(fail BOOL) RETURNS INTEGER AS $$
DECLARE
  v_var INTEGER;`,
			},
			{
				Statement: `BEGIN
  BEGIN
    v_var := (leaker_2(fail)).error_code;`,
			},
			{
				Statement: `  EXCEPTION
    WHEN others THEN RETURN 0;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `  RETURN 1;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE FUNCTION leaker_2(fail BOOL, OUT error_code INTEGER, OUT new_id INTEGER)
  RETURNS RECORD AS $$
BEGIN
  IF fail THEN
    RAISE EXCEPTION 'fail ...';`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `  error_code := 1;`,
			},
			{
				Statement: `  new_id := 1;`,
			},
			{
				Statement: `  RETURN;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT * FROM leaker_1(false);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM leaker_1(true);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `DROP FUNCTION leaker_1(bool);`,
			},
			{
				Statement: `DROP FUNCTION leaker_2(bool);`,
			},
			{
				Statement: `CREATE FUNCTION nonsimple_expr_test() RETURNS text[] AS $$
DECLARE
  arr text[];`,
			},
			{
				Statement: `  lr text;`,
			},
			{
				Statement: `  i integer;`,
			},
			{
				Statement: `BEGIN
  arr := array[array['foo','bar'], array['baz', 'quux']];`,
			},
			{
				Statement: `  lr := 'fool';`,
			},
			{
				Statement: `  i := 1;`,
			},
			{
				Statement: `  -- use sub-SELECTs to make expressions non-simple
  arr[(SELECT i)][(SELECT i+1)] := (SELECT lr);`,
			},
			{
				Statement: `  RETURN arr;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT nonsimple_expr_test();`,
				Results:   []sql.Row{{`{{foo,fool},{baz,quux}}`}},
			},
			{
				Statement: `DROP FUNCTION nonsimple_expr_test();`,
			},
			{
				Statement: `CREATE FUNCTION nonsimple_expr_test() RETURNS integer AS $$
declare
   i integer NOT NULL := 0;`,
			},
			{
				Statement: `begin
  begin
    i := (SELECT NULL::integer);  -- should throw error`,
			},
			{
				Statement: `  exception
    WHEN OTHERS THEN
      i := (SELECT 1::integer);`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  return i;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT nonsimple_expr_test();`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP FUNCTION nonsimple_expr_test();`,
			},
			{
				Statement: `create function recurse(float8) returns float8 as
$$
begin
  if ($1 > 0) then
    return sql_recurse($1 - 1);`,
			},
			{
				Statement: `  else
    return $1;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create function sql_recurse(float8) returns float8 as
$$ select recurse($1) limit 1; $$ language sql;`,
			},
			{
				Statement: `select recurse(10);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `create function error1(text) returns text language sql as
$$ SELECT relname::text FROM pg_class c WHERE c.oid = $1::regclass $$;`,
			},
			{
				Statement: `create function error2(p_name_table text) returns text language plpgsql as $$
begin
  return error1(p_name_table);`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `create table public.stuffs (stuff text);`,
			},
			{
				Statement: `SAVEPOINT a;`,
			},
			{
				Statement:   `select error2('nonexistent.stuffs');`,
				ErrorString: `schema "nonexistent" does not exist`,
			},
			{
				Statement: `CONTEXT:  SQL function "error1" statement 1
PL/pgSQL function error2(text) line 3 at RETURN
ROLLBACK TO a;`,
			},
			{
				Statement: `select error2('public.stuffs');`,
				Results:   []sql.Row{{`stuffs`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `drop function error2(p_name_table text);`,
			},
			{
				Statement: `drop function error1(text);`,
			},
			{
				Statement: `create function sql_to_date(integer) returns date as $$
select $1::text::date
$$ language sql immutable strict;`,
			},
			{
				Statement: `create cast (integer as date) with function sql_to_date(integer) as assignment;`,
			},
			{
				Statement: `create function cast_invoker(integer) returns date as $$
begin
  return $1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select cast_invoker(20150717);`,
				Results:   []sql.Row{{`07-17-2015`}},
			},
			{
				Statement: `select cast_invoker(20150718);  -- second call crashed in pre-release 9.5`,
				Results:   []sql.Row{{`07-18-2015`}},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `select cast_invoker(20150717);`,
				Results:   []sql.Row{{`07-17-2015`}},
			},
			{
				Statement: `select cast_invoker(20150718);`,
				Results:   []sql.Row{{`07-18-2015`}},
			},
			{
				Statement: `savepoint s1;`,
			},
			{
				Statement: `select cast_invoker(20150718);`,
				Results:   []sql.Row{{`07-18-2015`}},
			},
			{
				Statement:   `select cast_invoker(-1); -- fails`,
				ErrorString: `invalid input syntax for type date: "-1"`,
			},
			{
				Statement: `CONTEXT:  SQL function "sql_to_date" statement 1
PL/pgSQL function cast_invoker(integer) while casting return value to function's return type
rollback to savepoint s1;`,
			},
			{
				Statement: `select cast_invoker(20150719);`,
				Results:   []sql.Row{{`07-19-2015`}},
			},
			{
				Statement: `select cast_invoker(20150720);`,
				Results:   []sql.Row{{`07-20-2015`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `drop function cast_invoker(integer);`,
			},
			{
				Statement: `drop function sql_to_date(integer) cascade;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `do $$ declare x text[]; begin x := '{1.23, 4.56}'::numeric[]; end $$;`,
			},
			{
				Statement: `do $$ declare x text[]; begin x := '{1.23, 4.56}'::numeric[]; end $$;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `create function fail() returns int language plpgsql as $$
begin
  return 1/0;`,
			},
			{
				Statement: `end
$$;`,
			},
			{
				Statement:   `select fail();`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `CONTEXT:  SQL expression "1/0"
PL/pgSQL function fail() line 3 at RETURN
select fail();`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `CONTEXT:  SQL expression "1/0"
PL/pgSQL function fail() line 3 at RETURN
drop function fail();`,
			},
			{
				Statement: `set standard_conforming_strings = off;`,
			},
			{
				Statement: `create or replace function strtest() returns text as $$
begin
  raise notice 'foo\\bar\041baz';`,
			},
			{
				Statement: `  return 'foo\\bar\041baz';`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `LINE 3:   raise notice 'foo\\bar\041baz';`,
			},
			{
				Statement: `                       ^
HINT:  Use the escape string syntax for backslashes, e.g., E'\\'.
LINE 4:   return 'foo\\bar\041baz';`,
			},
			{
				Statement: `                 ^
HINT:  Use the escape string syntax for backslashes, e.g., E'\\'.
LINE 4:   return 'foo\\bar\041baz';`,
			},
			{
				Statement: `                 ^
HINT:  Use the escape string syntax for backslashes, e.g., E'\\'.
select strtest();`,
			},
			{
				Statement: `LINE 1: 'foo\\bar\041baz'
        ^
HINT:  Use the escape string syntax for backslashes, e.g., E'\\'.
QUERY:  'foo\\bar\041baz'
   strtest   
-------------
 foo\bar!baz
(1 row)
create or replace function strtest() returns text as $$
begin
  raise notice E'foo\\bar\041baz';`,
			},
			{
				Statement: `  return E'foo\\bar\041baz';`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select strtest();`,
				Results:   []sql.Row{{`foo\bar!baz`}},
			},
			{
				Statement: `set standard_conforming_strings = on;`,
			},
			{
				Statement: `create or replace function strtest() returns text as $$
begin
  raise notice 'foo\\bar\041baz\';`,
			},
			{
				Statement: `  return 'foo\\bar\041baz\';`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select strtest();`,
				Results:   []sql.Row{{`foo\\bar\041baz\`}},
			},
			{
				Statement: `create or replace function strtest() returns text as $$
begin
  raise notice E'foo\\bar\041baz';`,
			},
			{
				Statement: `  return E'foo\\bar\041baz';`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select strtest();`,
				Results:   []sql.Row{{`foo\bar!baz`}},
			},
			{
				Statement: `drop function strtest();`,
			},
			{
				Statement: `DO $$
DECLARE r record;`,
			},
			{
				Statement: `BEGIN
    FOR r IN SELECT rtrim(roomno) AS roomno, comment FROM Room ORDER BY roomno
    LOOP
        RAISE NOTICE '%, %', r.roomno, r.comment;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END$$;`,
			},
			{
				Statement:   `DO LANGUAGE plpgsql $$begin return 1; end$$;`,
				ErrorString: `RETURN cannot have a parameter in function returning void`,
			},
			{
				Statement: `DO $$
DECLARE r record;`,
			},
			{
				Statement: `BEGIN
    FOR r IN SELECT rtrim(roomno) AS roomno, foo FROM Room ORDER BY roomno
    LOOP
        RAISE NOTICE '%, %', r.roomno, r.comment;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement:   `END$$;`,
				ErrorString: `column "foo" does not exist`,
			},
			{
				Statement: `QUERY:  SELECT rtrim(roomno) AS roomno, foo FROM Room ORDER BY roomno
CONTEXT:  PL/pgSQL function inline_code_block line 4 at FOR over SELECT rows
do $outer$
begin
  for i in 1..10 loop
   begin
    execute $ex$
      do $$
      declare x int = 0;`,
			},
			{
				Statement: `      begin
        x := 1 / x;`,
			},
			{
				Statement: `      end;`,
			},
			{
				Statement: `      $$;`,
			},
			{
				Statement: `    $ex$;`,
			},
			{
				Statement: `  exception when division_by_zero then
    raise notice 'caught division by zero';`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$outer$;`,
			},
			{
				Statement: `do $$
declare x int := x + 1;  -- error`,
			},
			{
				Statement: `begin
  raise notice 'x = %', x;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `column "x" does not exist`,
			},
			{
				Statement: `QUERY:  x + 1
CONTEXT:  PL/pgSQL function inline_code_block line 2 during statement block local variable initialization
do $$
declare y int := x + 1;  -- error`,
			},
			{
				Statement: `        x int := 42;`,
			},
			{
				Statement: `begin
  raise notice 'x = %, y = %', x, y;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `column "x" does not exist`,
			},
			{
				Statement: `QUERY:  x + 1
CONTEXT:  PL/pgSQL function inline_code_block line 2 during statement block local variable initialization
do $$
declare x int := 42;`,
			},
			{
				Statement: `        y int := x + 1;`,
			},
			{
				Statement: `begin
  raise notice 'x = %, y = %', x, y;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `do $$
declare x int := 42;`,
			},
			{
				Statement: `begin
  declare y int := x + 1;`,
			},
			{
				Statement: `          x int := x + 2;`,
			},
			{
				Statement: `          z int := x * 10;`,
			},
			{
				Statement: `  begin
    raise notice 'x = %, y = %, z = %', x, y, z;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `set plpgsql.variable_conflict = error;`,
			},
			{
				Statement: `create function conflict_test() returns setof int8_tbl as $$
declare r record;`,
			},
			{
				Statement: `  q1 bigint := 42;`,
			},
			{
				Statement: `begin
  for r in select q1,q2 from int8_tbl loop
    return next r;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select * from conflict_test();`,
				ErrorString: `column reference "q1" is ambiguous`,
			},
			{
				Statement: `QUERY:  select q1,q2 from int8_tbl
CONTEXT:  PL/pgSQL function conflict_test() line 5 at FOR over SELECT rows
create or replace function conflict_test() returns setof int8_tbl as $$
#variable_conflict use_variable
declare r record;`,
			},
			{
				Statement: `  q1 bigint := 42;`,
			},
			{
				Statement: `begin
  for r in select q1,q2 from int8_tbl loop
    return next r;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from conflict_test();`,
				Results:   []sql.Row{{42, 456}, {42, 4567890123456789}, {42, 123}, {42, 4567890123456789}, {42, -4567890123456789}},
			},
			{
				Statement: `create or replace function conflict_test() returns setof int8_tbl as $$
#variable_conflict use_column
declare r record;`,
			},
			{
				Statement: `  q1 bigint := 42;`,
			},
			{
				Statement: `begin
  for r in select q1,q2 from int8_tbl loop
    return next r;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select * from conflict_test();`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `drop function conflict_test();`,
			},
			{
				Statement: `create function unreserved_test() returns int as $$
declare
  forward int := 21;`,
			},
			{
				Statement: `begin
  forward := forward * 2;`,
			},
			{
				Statement: `  return forward;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select unreserved_test();`,
				Results:   []sql.Row{{42}},
			},
			{
				Statement: `create or replace function unreserved_test() returns int as $$
declare
  return int := 42;`,
			},
			{
				Statement: `begin
  return := return + 1;`,
			},
			{
				Statement: `  return return;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select unreserved_test();`,
				Results:   []sql.Row{{43}},
			},
			{
				Statement: `create or replace function unreserved_test() returns int as $$
declare
  comment int := 21;`,
			},
			{
				Statement: `begin
  comment := comment * 2;`,
			},
			{
				Statement: `  comment on function unreserved_test() is 'this is a test';`,
			},
			{
				Statement: `  return comment;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select unreserved_test();`,
				Results:   []sql.Row{{42}},
			},
			{
				Statement: `select obj_description('unreserved_test()'::regprocedure, 'pg_proc');`,
				Results:   []sql.Row{{`this is a test`}},
			},
			{
				Statement: `drop function unreserved_test();`,
			},
			{
				Statement: `create function foreach_test(anyarray)
returns void as $$
declare x int;`,
			},
			{
				Statement: `begin
  foreach x in array $1
  loop
    raise notice '%', x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select foreach_test(ARRAY[1,2,3,4]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select foreach_test(ARRAY[[1,2],[3,4]]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function foreach_test(anyarray)
returns void as $$
declare x int;`,
			},
			{
				Statement: `begin
  foreach x slice 1 in array $1
  loop
    raise notice '%', x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select foreach_test(ARRAY[1,2,3,4]);`,
				ErrorString: `FOREACH ... SLICE loop variable must be of an array type`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function foreach_test(anyarray) line 4 at FOREACH over array
select foreach_test(ARRAY[[1,2],[3,4]]);`,
				ErrorString: `FOREACH ... SLICE loop variable must be of an array type`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function foreach_test(anyarray) line 4 at FOREACH over array
create or replace function foreach_test(anyarray)
returns void as $$
declare x int[];`,
			},
			{
				Statement: `begin
  foreach x slice 1 in array $1
  loop
    raise notice '%', x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select foreach_test(ARRAY[1,2,3,4]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select foreach_test(ARRAY[[1,2],[3,4]]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function foreach_test(anyarray)
returns void as $$
declare x int[];`,
			},
			{
				Statement: `begin
  foreach x slice 2 in array $1
  loop
    raise notice '%', x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement:   `select foreach_test(ARRAY[1,2,3,4]);`,
				ErrorString: `slice dimension (2) is out of the valid range 0..1`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function foreach_test(anyarray) line 4 at FOREACH over array
select foreach_test(ARRAY[[1,2],[3,4]]);`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `select foreach_test(ARRAY[[[1,2]],[[3,4]]]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create type xy_tuple AS (x int, y int);`,
			},
			{
				Statement: `create or replace function foreach_test(anyarray)
returns void as $$
declare r record;`,
			},
			{
				Statement: `begin
  foreach r in array $1
  loop
    raise notice '%', r;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select foreach_test(ARRAY[(10,20),(40,69),(35,78)]::xy_tuple[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select foreach_test(ARRAY[[(10,20),(40,69)],[(35,78),(88,76)]]::xy_tuple[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function foreach_test(anyarray)
returns void as $$
declare x int; y int;`,
			},
			{
				Statement: `begin
  foreach x, y in array $1
  loop
    raise notice 'x = %, y = %', x, y;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select foreach_test(ARRAY[(10,20),(40,69),(35,78)]::xy_tuple[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select foreach_test(ARRAY[[(10,20),(40,69)],[(35,78),(88,76)]]::xy_tuple[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create or replace function foreach_test(anyarray)
returns void as $$
declare x xy_tuple[];`,
			},
			{
				Statement: `begin
  foreach x slice 1 in array $1
  loop
    raise notice '%', x;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select foreach_test(ARRAY[(10,20),(40,69),(35,78)]::xy_tuple[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select foreach_test(ARRAY[[(10,20),(40,69)],[(35,78),(88,76)]]::xy_tuple[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function foreach_test(anyarray);`,
			},
			{
				Statement: `drop type xy_tuple;`,
			},
			{
				Statement: `create temp table rtype (id int, ar text[]);`,
			},
			{
				Statement: `create function arrayassign1() returns text[] language plpgsql as $$
declare
 r record;`,
			},
			{
				Statement: `begin
  r := row(12, '{foo,bar,baz}')::rtype;`,
			},
			{
				Statement: `  r.ar[2] := 'replace';`,
			},
			{
				Statement: `  return r.ar;`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `select arrayassign1();`,
				Results:   []sql.Row{{`{foo,replace,baz}`}},
			},
			{
				Statement: `select arrayassign1(); -- try again to exercise internal caching`,
				Results:   []sql.Row{{`{foo,replace,baz}`}},
			},
			{
				Statement: `create domain orderedarray as int[2]
  constraint sorted check (value[1] < value[2]);`,
			},
			{
				Statement: `select '{1,2}'::orderedarray;`,
				Results:   []sql.Row{{`{1,2}`}},
			},
			{
				Statement:   `select '{2,1}'::orderedarray;  -- fail`,
				ErrorString: `value for domain orderedarray violates check constraint "sorted"`,
			},
			{
				Statement: `create function testoa(x1 int, x2 int, x3 int) returns orderedarray
language plpgsql as $$
declare res orderedarray;`,
			},
			{
				Statement: `begin
  res := array[x1, x2];`,
			},
			{
				Statement: `  res[2] := x3;`,
			},
			{
				Statement: `  return res;`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `select testoa(1,2,3);`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement: `select testoa(1,2,3); -- try again to exercise internal caching`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement:   `select testoa(2,1,3); -- fail at initial assign`,
				ErrorString: `value for domain orderedarray violates check constraint "sorted"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function testoa(integer,integer,integer) line 4 at assignment
select testoa(1,2,1); -- fail at update`,
				ErrorString: `value for domain orderedarray violates check constraint "sorted"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function testoa(integer,integer,integer) line 5 at assignment
drop function arrayassign1();`,
			},
			{
				Statement: `drop function testoa(x1 int, x2 int, x3 int);`,
			},
			{
				Statement: `create function returns_rw_array(int) returns int[]
language plpgsql as $$
  declare r int[];`,
			},
			{
				Statement: `  begin r := array[$1, $1]; return r; end;`,
			},
			{
				Statement: `$$ stable;`,
			},
			{
				Statement: `create function consumes_rw_array(int[]) returns int
language plpgsql as $$
  begin return $1[1]; end;`,
			},
			{
				Statement: `$$ stable;`,
			},
			{
				Statement: `select consumes_rw_array(returns_rw_array(42));`,
				Results:   []sql.Row{{42}},
			},
			{
				Statement: `explain (verbose, costs off)
select i, a from
  (select returns_rw_array(1) as a offset 0) ss,
  lateral consumes_rw_array(a) i;`,
				Results: []sql.Row{{`Nested Loop`}, {`Output: i.i, (returns_rw_array(1))`}, {`->  Result`}, {`Output: returns_rw_array(1)`}, {`->  Function Scan on public.consumes_rw_array i`}, {`Output: i.i`}, {`Function Call: consumes_rw_array((returns_rw_array(1)))`}},
			},
			{
				Statement: `select i, a from
  (select returns_rw_array(1) as a offset 0) ss,
  lateral consumes_rw_array(a) i;`,
				Results: []sql.Row{{1, `{1,1}`}},
			},
			{
				Statement: `explain (verbose, costs off)
select consumes_rw_array(a), a from returns_rw_array(1) a;`,
				Results: []sql.Row{{`Function Scan on public.returns_rw_array a`}, {`Output: consumes_rw_array(a), a`}, {`Function Call: returns_rw_array(1)`}},
			},
			{
				Statement: `select consumes_rw_array(a), a from returns_rw_array(1) a;`,
				Results:   []sql.Row{{1, `{1,1}`}},
			},
			{
				Statement: `explain (verbose, costs off)
select consumes_rw_array(a), a from
  (values (returns_rw_array(1)), (returns_rw_array(2))) v(a);`,
				Results: []sql.Row{{`Values Scan on "*VALUES*"`}, {`Output: consumes_rw_array("*VALUES*".column1), "*VALUES*".column1`}},
			},
			{
				Statement: `select consumes_rw_array(a), a from
  (values (returns_rw_array(1)), (returns_rw_array(2))) v(a);`,
				Results: []sql.Row{{1, `{1,1}`}, {2, `{2,2}`}},
			},
			{
				Statement: `do $$
declare a int[] := array[1,2];`,
			},
			{
				Statement: `begin
  a := a || 3;`,
			},
			{
				Statement: `  raise notice 'a = %', a;`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `create function inner_func(int)
returns int as $$
declare _context text;`,
			},
			{
				Statement: `begin
  get diagnostics _context = pg_context;`,
			},
			{
				Statement: `  raise notice '***%***', _context;`,
			},
			{
				Statement: `  -- lets do it again, just for fun..
  get diagnostics _context = pg_context;`,
			},
			{
				Statement: `  raise notice '***%***', _context;`,
			},
			{
				Statement: `  raise notice 'lets make sure we didnt break anything';`,
			},
			{
				Statement: `  return 2 * $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create or replace function outer_func(int)
returns int as $$
declare
  myresult int;`,
			},
			{
				Statement: `begin
  raise notice 'calling down into inner_func()';`,
			},
			{
				Statement: `  myresult := inner_func($1);`,
			},
			{
				Statement: `  raise notice 'inner_func() done';`,
			},
			{
				Statement: `  return myresult;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create or replace function outer_outer_func(int)
returns int as $$
declare
  myresult int;`,
			},
			{
				Statement: `begin
  raise notice 'calling down into outer_func()';`,
			},
			{
				Statement: `  myresult := outer_func($1);`,
			},
			{
				Statement: `  raise notice 'outer_func() done';`,
			},
			{
				Statement: `  return myresult;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select outer_outer_func(10);`,
			},
			{
				Statement: `PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
 outer_outer_func 
------------------
               20
(1 row)
select outer_outer_func(20);`,
			},
			{
				Statement: `PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
 outer_outer_func 
------------------
               40
(1 row)
drop function outer_outer_func(int);`,
			},
			{
				Statement: `drop function outer_func(int);`,
			},
			{
				Statement: `drop function inner_func(int);`,
			},
			{
				Statement: `create function inner_func(int)
returns int as $$
declare
  _context text;`,
			},
			{
				Statement: `  sx int := 5;`,
			},
			{
				Statement: `begin
  begin
    perform sx / 0;`,
			},
			{
				Statement: `  exception
    when division_by_zero then
      get diagnostics _context = pg_context;`,
			},
			{
				Statement: `      raise notice '***%***', _context;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  -- lets do it again, just for fun..
  get diagnostics _context = pg_context;`,
			},
			{
				Statement: `  raise notice '***%***', _context;`,
			},
			{
				Statement: `  raise notice 'lets make sure we didnt break anything';`,
			},
			{
				Statement: `  return 2 * $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create or replace function outer_func(int)
returns int as $$
declare
  myresult int;`,
			},
			{
				Statement: `begin
  raise notice 'calling down into inner_func()';`,
			},
			{
				Statement: `  myresult := inner_func($1);`,
			},
			{
				Statement: `  raise notice 'inner_func() done';`,
			},
			{
				Statement: `  return myresult;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create or replace function outer_outer_func(int)
returns int as $$
declare
  myresult int;`,
			},
			{
				Statement: `begin
  raise notice 'calling down into outer_func()';`,
			},
			{
				Statement: `  myresult := outer_func($1);`,
			},
			{
				Statement: `  raise notice 'outer_func() done';`,
			},
			{
				Statement: `  return myresult;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `select outer_outer_func(10);`,
			},
			{
				Statement: `PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
 outer_outer_func 
------------------
               20
(1 row)
select outer_outer_func(20);`,
			},
			{
				Statement: `PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
PL/pgSQL function outer_func(integer) line 6 at assignment
PL/pgSQL function outer_outer_func(integer) line 6 at assignment***
 outer_outer_func 
------------------
               40
(1 row)
drop function outer_outer_func(int);`,
			},
			{
				Statement: `drop function outer_func(int);`,
			},
			{
				Statement: `drop function inner_func(int);`,
			},
			{
				Statement: `do $$
begin
  assert 1=1;  -- should succeed`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `do $$
begin
  assert 1=0;  -- should fail`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `assertion failed`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 3 at ASSERT
do $$
begin
  assert NULL;  -- should fail`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `assertion failed`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 3 at ASSERT
set plpgsql.check_asserts = off;`,
			},
			{
				Statement: `do $$
begin
  assert 1=0;  -- won't be tested`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `reset plpgsql.check_asserts;`,
			},
			{
				Statement: `do $$
declare var text := 'some value';`,
			},
			{
				Statement: `begin
  assert 1=0, format('assertion failed, var = "%s"', var);`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `assertion failed, var = "some value"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 4 at ASSERT
do $$
begin
  assert 1=0, 'unhandled assertion';`,
			},
			{
				Statement: `exception when others then
  null; -- do nothing`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `unhandled assertion`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 3 at ASSERT
create function plpgsql_domain_check(val int) returns boolean as $$
begin return val > 0; end
$$ language plpgsql immutable;`,
			},
			{
				Statement: `create domain plpgsql_domain as integer check(plpgsql_domain_check(value));`,
			},
			{
				Statement: `do $$
declare v_test plpgsql_domain;`,
			},
			{
				Statement: `begin
  v_test := 1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `do $$
declare v_test plpgsql_domain := 1;`,
			},
			{
				Statement: `begin
  v_test := 0;  -- fail`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `value for domain plpgsql_domain violates check constraint "plpgsql_domain_check"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 4 at assignment
create function plpgsql_arr_domain_check(val int[]) returns boolean as $$
begin return val[1] > 0; end
$$ language plpgsql immutable;`,
			},
			{
				Statement: `create domain plpgsql_arr_domain as int[] check(plpgsql_arr_domain_check(value));`,
			},
			{
				Statement: `do $$
declare v_test plpgsql_arr_domain;`,
			},
			{
				Statement: `begin
  v_test := array[1];`,
			},
			{
				Statement: `  v_test := v_test || 2;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `do $$
declare v_test plpgsql_arr_domain := array[1];`,
			},
			{
				Statement: `begin
  v_test := 0 || v_test;  -- fail`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement:   `$$;`,
				ErrorString: `value for domain plpgsql_arr_domain violates check constraint "plpgsql_arr_domain_check"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function inline_code_block line 4 at assignment
CREATE TABLE transition_table_base (id int PRIMARY KEY, val text);`,
			},
			{
				Statement: `CREATE FUNCTION transition_table_base_ins_func()
  RETURNS trigger
  LANGUAGE plpgsql
AS $$
DECLARE
  t text;`,
			},
			{
				Statement: `  l text;`,
			},
			{
				Statement: `BEGIN
  t = '';`,
			},
			{
				Statement: `  FOR l IN EXECUTE
           $q$
             EXPLAIN (TIMING off, COSTS off, VERBOSE on)
             SELECT * FROM newtable
           $q$ LOOP
    t = t || l || E'\n';`,
			},
			{
				Statement: `  END LOOP;`,
			},
			{
				Statement: `  RAISE INFO '%', t;`,
			},
			{
				Statement: `  RETURN new;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_base_ins_trig
  AFTER INSERT ON transition_table_base
  REFERENCING OLD TABLE AS oldtable NEW TABLE AS newtable
  FOR EACH STATEMENT
  EXECUTE PROCEDURE transition_table_base_ins_func();`,
				ErrorString: `OLD TABLE can only be specified for a DELETE or UPDATE trigger`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_base_ins_trig
  AFTER INSERT ON transition_table_base
  REFERENCING NEW TABLE AS newtable
  FOR EACH STATEMENT
  EXECUTE PROCEDURE transition_table_base_ins_func();`,
			},
			{
				Statement: `INSERT INTO transition_table_base VALUES (1, 'One'), (2, 'Two');`,
			},
			{
				Statement: `INFO:  Named Tuplestore Scan
  Output: id, val
INSERT INTO transition_table_base VALUES (3, 'Three'), (4, 'Four');`,
			},
			{
				Statement: `INFO:  Named Tuplestore Scan
  Output: id, val
CREATE OR REPLACE FUNCTION transition_table_base_upd_func()
  RETURNS trigger
  LANGUAGE plpgsql
AS $$
DECLARE
  t text;`,
			},
			{
				Statement: `  l text;`,
			},
			{
				Statement: `BEGIN
  t = '';`,
			},
			{
				Statement: `  FOR l IN EXECUTE
           $q$
             EXPLAIN (TIMING off, COSTS off, VERBOSE on)
             SELECT * FROM oldtable ot FULL JOIN newtable nt USING (id)
           $q$ LOOP
    t = t || l || E'\n';`,
			},
			{
				Statement: `  END LOOP;`,
			},
			{
				Statement: `  RAISE INFO '%', t;`,
			},
			{
				Statement: `  RETURN new;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_base_upd_trig
  AFTER UPDATE ON transition_table_base
  REFERENCING OLD TABLE AS oldtable NEW TABLE AS newtable
  FOR EACH STATEMENT
  EXECUTE PROCEDURE transition_table_base_upd_func();`,
			},
			{
				Statement: `UPDATE transition_table_base
  SET val = '*' || val || '*'
  WHERE id BETWEEN 2 AND 3;`,
			},
			{
				Statement: `INFO:  Hash Full Join
  Output: COALESCE(ot.id, nt.id), ot.val, nt.val
  Hash Cond: (ot.id = nt.id)
  ->  Named Tuplestore Scan
        Output: ot.id, ot.val
  ->  Hash
        Output: nt.id, nt.val
        ->  Named Tuplestore Scan
              Output: nt.id, nt.val
CREATE TABLE transition_table_level1
(
      level1_no serial NOT NULL ,
      level1_node_name varchar(255),
       PRIMARY KEY (level1_no)
) WITHOUT OIDS;`,
			},
			{
				Statement: `CREATE TABLE transition_table_level2
(
      level2_no serial NOT NULL ,
      parent_no int NOT NULL,
      level1_node_name varchar(255),
       PRIMARY KEY (level2_no)
) WITHOUT OIDS;`,
			},
			{
				Statement: `CREATE TABLE transition_table_status
(
      level int NOT NULL,
      node_no int NOT NULL,
      status int,
       PRIMARY KEY (level, node_no)
) WITHOUT OIDS;`,
			},
			{
				Statement: `CREATE FUNCTION transition_table_level1_ri_parent_del_func()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
  DECLARE n bigint;`,
			},
			{
				Statement: `  BEGIN
    PERFORM FROM p JOIN transition_table_level2 c ON c.parent_no = p.level1_no;`,
			},
			{
				Statement: `    IF FOUND THEN
      RAISE EXCEPTION 'RI error';`,
			},
			{
				Statement: `    END IF;`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_level1_ri_parent_del_trigger
  AFTER DELETE ON transition_table_level1
  REFERENCING OLD TABLE AS p
  FOR EACH STATEMENT EXECUTE PROCEDURE
    transition_table_level1_ri_parent_del_func();`,
			},
			{
				Statement: `CREATE FUNCTION transition_table_level1_ri_parent_upd_func()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
  DECLARE
    x int;`,
			},
			{
				Statement: `  BEGIN
    WITH p AS (SELECT level1_no, sum(delta) cnt
                 FROM (SELECT level1_no, 1 AS delta FROM i
                       UNION ALL
                       SELECT level1_no, -1 AS delta FROM d) w
                 GROUP BY level1_no
                 HAVING sum(delta) < 0)
    SELECT level1_no
      FROM p JOIN transition_table_level2 c ON c.parent_no = p.level1_no
      INTO x;`,
			},
			{
				Statement: `    IF FOUND THEN
      RAISE EXCEPTION 'RI error';`,
			},
			{
				Statement: `    END IF;`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_level1_ri_parent_upd_trigger
  AFTER UPDATE ON transition_table_level1
  REFERENCING OLD TABLE AS d NEW TABLE AS i
  FOR EACH STATEMENT EXECUTE PROCEDURE
    transition_table_level1_ri_parent_upd_func();`,
			},
			{
				Statement: `CREATE FUNCTION transition_table_level2_ri_child_insupd_func()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
  BEGIN
    PERFORM FROM i
      LEFT JOIN transition_table_level1 p
        ON p.level1_no IS NOT NULL AND p.level1_no = i.parent_no
      WHERE p.level1_no IS NULL;`,
			},
			{
				Statement: `    IF FOUND THEN
      RAISE EXCEPTION 'RI error';`,
			},
			{
				Statement: `    END IF;`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_level2_ri_child_ins_trigger
  AFTER INSERT ON transition_table_level2
  REFERENCING NEW TABLE AS i
  FOR EACH STATEMENT EXECUTE PROCEDURE
    transition_table_level2_ri_child_insupd_func();`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_level2_ri_child_upd_trigger
  AFTER UPDATE ON transition_table_level2
  REFERENCING NEW TABLE AS i
  FOR EACH STATEMENT EXECUTE PROCEDURE
    transition_table_level2_ri_child_insupd_func();`,
			},
			{
				Statement: `INSERT INTO transition_table_level1 (level1_no)
  SELECT generate_series(1,200);`,
			},
			{
				Statement: `ANALYZE transition_table_level1;`,
			},
			{
				Statement: `INSERT INTO transition_table_level2 (level2_no, parent_no)
  SELECT level2_no, level2_no / 50 + 1 AS parent_no
    FROM generate_series(1,9999) level2_no;`,
			},
			{
				Statement: `ANALYZE transition_table_level2;`,
			},
			{
				Statement: `INSERT INTO transition_table_status (level, node_no, status)
  SELECT 1, level1_no, 0 FROM transition_table_level1;`,
			},
			{
				Statement: `INSERT INTO transition_table_status (level, node_no, status)
  SELECT 2, level2_no, 0 FROM transition_table_level2;`,
			},
			{
				Statement: `ANALYZE transition_table_status;`,
			},
			{
				Statement: `INSERT INTO transition_table_level1(level1_no)
  SELECT generate_series(201,1000);`,
			},
			{
				Statement: `ANALYZE transition_table_level1;`,
			},
			{
				Statement: `CREATE FUNCTION transition_table_level2_bad_usage_func()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
  BEGIN
    INSERT INTO dx VALUES (1000000, 1000000, 'x');`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER transition_table_level2_bad_usage_trigger
  AFTER DELETE ON transition_table_level2
  REFERENCING OLD TABLE AS dx
  FOR EACH STATEMENT EXECUTE PROCEDURE
    transition_table_level2_bad_usage_func();`,
			},
			{
				Statement: `DELETE FROM transition_table_level2
  WHERE level2_no BETWEEN 301 AND 305;`,
				ErrorString: `relation "dx" cannot be the target of a modifying statement`,
			},
			{
				Statement: `CONTEXT:  SQL statement "INSERT INTO dx VALUES (1000000, 1000000, 'x')"
PL/pgSQL function transition_table_level2_bad_usage_func() line 3 at SQL statement
DROP TRIGGER transition_table_level2_bad_usage_trigger
  ON transition_table_level2;`,
			},
			{
				Statement: `DELETE FROM transition_table_level1
  WHERE level1_no = 25;`,
				ErrorString: `RI error`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function transition_table_level1_ri_parent_del_func() line 6 at RAISE
UPDATE transition_table_level1 SET level1_no = -1
  WHERE level1_no = 30;`,
				ErrorString: `RI error`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function transition_table_level1_ri_parent_upd_func() line 15 at RAISE
INSERT INTO transition_table_level2 (level2_no, parent_no)
  VALUES (10000, 10000);`,
				ErrorString: `RI error`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function transition_table_level2_ri_child_insupd_func() line 8 at RAISE
UPDATE transition_table_level2 SET parent_no = 2000
  WHERE level2_no = 40;`,
				ErrorString: `RI error`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function transition_table_level2_ri_child_insupd_func() line 8 at RAISE
DELETE FROM transition_table_level1
  WHERE level1_no BETWEEN 201 AND 1000;`,
			},
			{
				Statement: `DELETE FROM transition_table_level1
  WHERE level1_no BETWEEN 100000000 AND 100000010;`,
			},
			{
				Statement: `SELECT count(*) FROM transition_table_level1;`,
				Results:   []sql.Row{{200}},
			},
			{
				Statement: `DELETE FROM transition_table_level2
  WHERE level2_no BETWEEN 211 AND 220;`,
			},
			{
				Statement: `SELECT count(*) FROM transition_table_level2;`,
				Results:   []sql.Row{{9989}},
			},
			{
				Statement: `CREATE TABLE alter_table_under_transition_tables
(
  id int PRIMARY KEY,
  name text
);`,
			},
			{
				Statement: `CREATE FUNCTION alter_table_under_transition_tables_upd_func()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  RAISE WARNING 'old table = %, new table = %',
                  (SELECT string_agg(id || '=' || name, ',') FROM d),
                  (SELECT string_agg(id || '=' || name, ',') FROM i);`,
			},
			{
				Statement: `  RAISE NOTICE 'one = %', (SELECT 1 FROM alter_table_under_transition_tables LIMIT 1);`,
			},
			{
				Statement: `  RETURN NULL;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER alter_table_under_transition_tables_upd_trigger
  AFTER TRUNCATE OR UPDATE ON alter_table_under_transition_tables
  REFERENCING OLD TABLE AS d NEW TABLE AS i
  FOR EACH STATEMENT EXECUTE PROCEDURE
    alter_table_under_transition_tables_upd_func();`,
				ErrorString: `TRUNCATE triggers with transition tables are not supported`,
			},
			{
				Statement: `CREATE TRIGGER alter_table_under_transition_tables_upd_trigger
  AFTER UPDATE ON alter_table_under_transition_tables
  REFERENCING OLD TABLE AS d NEW TABLE AS i
  FOR EACH STATEMENT EXECUTE PROCEDURE
    alter_table_under_transition_tables_upd_func();`,
			},
			{
				Statement: `INSERT INTO alter_table_under_transition_tables
  VALUES (1, '1'), (2, '2'), (3, '3');`,
			},
			{
				Statement: `UPDATE alter_table_under_transition_tables
  SET name = name || name;`,
			},
			{
				Statement: `ALTER TABLE alter_table_under_transition_tables
  ALTER COLUMN name TYPE int USING name::integer;`,
			},
			{
				Statement: `UPDATE alter_table_under_transition_tables
  SET name = (name::text || name::text)::integer;`,
			},
			{
				Statement: `ALTER TABLE alter_table_under_transition_tables
  DROP column name;`,
			},
			{
				Statement: `UPDATE alter_table_under_transition_tables
  SET id = id;`,
				ErrorString: `column "name" does not exist`,
			},
			{
				Statement: `QUERY:  (SELECT string_agg(id || '=' || name, ',') FROM d)
CONTEXT:  PL/pgSQL function alter_table_under_transition_tables_upd_func() line 3 at RAISE
CREATE TABLE multi_test (i int);`,
			},
			{
				Statement: `INSERT INTO multi_test VALUES (1);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION multi_test_trig() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE NOTICE 'count = %', (SELECT COUNT(*) FROM new_test);`,
			},
			{
				Statement: `    RAISE NOTICE 'count union = %',
      (SELECT COUNT(*)
       FROM (SELECT * FROM new_test UNION ALL SELECT * FROM new_test) ss);`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `END$$;`,
			},
			{
				Statement: `CREATE TRIGGER my_trigger AFTER UPDATE ON multi_test
  REFERENCING NEW TABLE AS new_test OLD TABLE as old_test
  FOR EACH STATEMENT EXECUTE PROCEDURE multi_test_trig();`,
			},
			{
				Statement: `UPDATE multi_test SET i = i;`,
			},
			{
				Statement: `DROP TABLE multi_test;`,
			},
			{
				Statement: `DROP FUNCTION multi_test_trig();`,
			},
			{
				Statement: `CREATE TABLE partitioned_table (a int, b text) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE pt_part1 PARTITION OF partitioned_table FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE pt_part2 PARTITION OF partitioned_table FOR VALUES IN (2);`,
			},
			{
				Statement: `INSERT INTO partitioned_table VALUES (1, 'Row 1');`,
			},
			{
				Statement: `INSERT INTO partitioned_table VALUES (2, 'Row 2');`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION get_from_partitioned_table(partitioned_table.a%type)
RETURNS partitioned_table AS $$
DECLARE
    a_val partitioned_table.a%TYPE;`,
			},
			{
				Statement: `    result partitioned_table%ROWTYPE;`,
			},
			{
				Statement: `BEGIN
    a_val := $1;`,
			},
			{
				Statement: `    SELECT * INTO result FROM partitioned_table WHERE a = a_val;`,
			},
			{
				Statement: `    RETURN result;`,
			},
			{
				Statement: `END; $$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT * FROM get_from_partitioned_table(1) AS t;`,
				Results:   []sql.Row{{1, `Row 1`}},
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION list_partitioned_table()
RETURNS SETOF partitioned_table.a%TYPE AS $$
DECLARE
    row partitioned_table%ROWTYPE;`,
			},
			{
				Statement: `    a_val partitioned_table.a%TYPE;`,
			},
			{
				Statement: `BEGIN
    FOR row IN SELECT * FROM partitioned_table ORDER BY a LOOP
        a_val := row.a;`,
			},
			{
				Statement: `        RETURN NEXT a_val;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `    RETURN;`,
			},
			{
				Statement: `END; $$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT * FROM list_partitioned_table() AS t;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `CREATE FUNCTION fx(x WSlot) RETURNS void AS $$
BEGIN
  GET DIAGNOSTICS x = ROW_COUNT;`,
			},
			{
				Statement: `  RETURN;`,
			},
			{
				Statement:   `END; $$ LANGUAGE plpgsql;`,
				ErrorString: `"x" is not a scalar variable`,
			},
		},
	})
}
