-- AUTOGENERATED BY gopkg.in/spacemonkeygo/dbx.v1
-- DO NOT EDIT
CREATE TABLE addresses (
	pk bigserial NOT NULL,
	buyer_pk bigint NOT NULL,
	created_at timestamp with time zone NOT NULL,
	street_address text NOT NULL,
	city text NOT NULL,
	state text NOT NULL,
	zip integer NOT NULL,
	is_billing boolean NOT NULL,
	id text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE buyers (
	pk bigserial NOT NULL,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL,
	id text NOT NULL,
	first_name text NOT NULL,
	last_name text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE buyer_emails (
	pk bigserial NOT NULL,
	buyer_pk bigint NOT NULL,
	created_at timestamp with time zone NOT NULL,
	address text NOT NULL,
	salted_hash text NOT NULL,
	id text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id ),
	UNIQUE ( address )
);
CREATE TABLE buyer_sessions (
	pk bigserial NOT NULL,
	buyer_pk bigint NOT NULL,
	id text NOT NULL,
	created_at timestamp with time zone NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE conversations (
	pk bigserial NOT NULL,
	vendor_pk bigint NOT NULL,
	buyer_pk bigint NOT NULL,
	buyer_unread boolean NOT NULL,
	vendor_unread boolean NOT NULL,
	message_count bigint NOT NULL,
	id text NOT NULL,
	created_at timestamp with time zone NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE executive_contacts (
	pk bigserial NOT NULL,
	id text NOT NULL,
	vendor_pk bigint NOT NULL,
	first_name text NOT NULL,
	last_name text NOT NULL,
	created_at timestamp with time zone NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE messages (
	pk bigserial NOT NULL,
	id text NOT NULL,
	created_at timestamp with time zone NOT NULL,
	buyer_sent boolean NOT NULL,
	description text NOT NULL,
	conversation_pk bigint NOT NULL,
	conversation_number bigint NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE products (
	pk bigserial NOT NULL,
	id text NOT NULL,
	vendor_pk bigint NOT NULL,
	created_at timestamp with time zone NOT NULL,
	price real NOT NULL,
	discount real NOT NULL,
	discount_active boolean NOT NULL,
	sku text NOT NULL,
	google_bucket_id text NOT NULL,
	ladybug_approved boolean NOT NULL,
	product_active boolean NOT NULL,
	num_in_stock integer NOT NULL,
	description text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE purchased_products (
	pk bigserial NOT NULL,
	id text NOT NULL,
	vendor_pk bigint NOT NULL,
	buyer_pk bigint NOT NULL,
	product_pk bigint NOT NULL,
	purchase_price real NOT NULL,
	created_at timestamp with time zone NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE trial_products (
	pk bigserial NOT NULL,
	id text NOT NULL,
	vendor_pk bigint NOT NULL,
	buyer_pk bigint NOT NULL,
	product_pk bigint NOT NULL,
	created_at timestamp with time zone NOT NULL,
	trial_price real NOT NULL,
	is_returned boolean NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE vendors (
	pk bigserial NOT NULL,
	id text NOT NULL,
	created_at timestamp with time zone NOT NULL,
	fein text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE vendor_addresses (
	pk bigserial NOT NULL,
	vendor_pk bigint NOT NULL,
	created_at timestamp with time zone NOT NULL,
	street_address text NOT NULL,
	city text NOT NULL,
	state text NOT NULL,
	zip integer NOT NULL,
	is_billing boolean NOT NULL,
	id text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
CREATE TABLE vendor_emails (
	pk bigserial NOT NULL,
	id text NOT NULL,
	executive_contact_pk bigint NOT NULL,
	created_at timestamp with time zone NOT NULL,
	address text NOT NULL,
	salted_hash text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id ),
	UNIQUE ( address )
);
CREATE TABLE vendor_phones (
	pk bigserial NOT NULL,
	id text NOT NULL,
	executive_contact_pk bigint NOT NULL,
	phone_number integer NOT NULL,
	country_code integer NOT NULL,
	area_code integer NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id ),
	UNIQUE ( phone_number )
);
CREATE TABLE vendor_sessions (
	pk bigserial NOT NULL,
	vendor_pk bigint NOT NULL,
	id text NOT NULL,
	created_at timestamp with time zone NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
);
