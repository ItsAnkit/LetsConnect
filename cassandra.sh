echo "Creating Keyspace......"
echo "CREATE KEYSPACE letsconnect
    WITH REPLICATION = {
        'class': 'SimpleStrategy', 'replication_factor': 1
    };" | cqlsh --cqlversion 3.4.2


echo "Creating Tables......"
echo "
use letsconnect;
create table users (
    id int,
    mobile_no int,
    username text,
    created_at timestamp,
    primary key(id)
);" | cqlsh --cqlversion 3.4.2

echo

echo "
use letsconnect;
create table conversations (
    id int primary key,
    creator_id int,
    created_at timestamp
);" | cqlsh --cqlversion 3.4.2

echo "
use letsconnect;
create table participants (
    id int primary key,
    conversation_id int,
    user_id int,
    created_at timestamp
);" | cqlsh --cqlversion 3.4.2

echo "
use letsconnect;
create table messages (
    id int primary key,
    sender_id int,
    conversation_id int,
    message text,
    created_at timestamp
);" | cqlsh --cqlversion 3.4.2

echo "Done !!!"
