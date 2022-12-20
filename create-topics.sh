echo -e 'Waiting for kafka'
echo -e 'Current topics:'
kafka-topics --bootstrap-server kafka:9092 --list

echo -e 'Creating kafka topics (if necessary)'
topics='user.registration.req user.login.req user.info.req user.registrations user.logins user.infos
ingredients.new ingredients.req ingredients'
for topic in $topics; do
    kafka-topics --bootstrap-server kafka:9092 --create --if-not-exists --topic "$topic" --replication-factor 1 --partitions 1
done

echo -e 'Topic list:'
kafka-topics --bootstrap-server kafka:9092 --list