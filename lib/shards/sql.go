package shards

const SqlCreateShardTable = `CREATE TABLE IF NOT EXISTS Shard (
	Address		VARCHAR(255) PRIMARY KEY
);`

const SqlCreateShardsMappingTable = `CREATE TABLE IF NOT EXISTS ShardsMapping (
	UserId				VARCHAR(255) PRIMARY KEY,
	ShardAddress		VARCHAR(255) REFERENCES Shard(Address)
);`

const SqlSelectShardByUser = `SELECT ShardAddress FROM ShardsMapping WHERE UserId = $1;`

const SqlDeleteUserShard = "DELETE FROM ShardsMapping WHERE UserId = $1;"

const SqlCreateUserShard = "INSERT INTO ShardsMapping (UserId, ShardAddress) VALUES ($1, $2);"

const SqlEnsureShard = `INSERT INTO Shard(Address) VALUES ($1) ON CONFLICT DO NOTHING;`

const SqlShardUserCount = `SELECT COUNT(ShardsMapping.UserId), Shard.Address
	FROM Shard LEFT JOIN ShardsMapping ON Shard.Address = ShardsMapping.ShardAddress
	GROUP BY Shard.Address;`

const SQLListShards = `SELECT Address FROM Shard`
