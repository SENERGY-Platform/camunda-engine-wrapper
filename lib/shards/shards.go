package shards

import (
	"context"
	"database/sql"
	"errors"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	_ "github.com/lib/pq"
	"time"
)

func New(pgConnStr string, cache cache.Cache) (*Shards, error) {
	db, err := initDbConnection(pgConnStr)
	if err != nil {
		return nil, err
	}
	return &Shards{db: db, cache: cache}, nil
}

func initDbConnection(conStr string) (db *sql.DB, err error) {
	db, err = sql.Open("postgres", conStr)
	if err != nil {
		return
	}
	_, err = db.Exec(SqlCreateShardTable)
	if err != nil {
		return db, err
	}
	_, err = db.Exec(SqlCreateShardsMappingTable)
	if err != nil {
		return db, err
	}
	return db, err
}

type Shards struct {
	db    *sql.DB
	cache cache.Cache
}

var ErrorNotFound = errors.New("no shard assigned to user")

const CachePrefix = "user-shard."

func (this *Shards) GetShardForUser(userId string) (shardUrl string, err error) {
	err = this.cache.Use(CachePrefix+userId, func() (interface{}, error) {
		return getShardForUser(this.db, userId)
	}, &shardUrl)
	return
}

func getShardForUser(tx Tx, userId string) (shardUrl string, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	resp := tx.QueryRowContext(ctx, SqlSelectShardByUser, userId)
	err = resp.Err()
	if err != nil {
		return
	}
	err = resp.Scan(&shardUrl)
	if err == sql.ErrNoRows {
		err = ErrorNotFound
	}
	return
}

func (this *Shards) SetShardForUser(userId string, shardAddress string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	tx, err := this.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	err = removeShardForUser(tx, userId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = addShardForUser(tx, userId, shardAddress)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return
	}
	return this.cache.Invalidate(CachePrefix + userId)
}

func (this *Shards) EnsureShardForUser(userId string) (shardUrl string, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	tx, err := this.db.BeginTx(ctx, nil)
	if err != nil {
		return shardUrl, err
	}

	err = this.cache.Use(CachePrefix+userId, func() (interface{}, error) {
		return getShardForUser(tx, userId)
	}, &shardUrl)

	//more work is only necessary if no shard is assigned to the user
	if err != ErrorNotFound {
		tx.Commit() //commit even if nothing changed to free locks
		return
	}
	shardUrl, err = selectShard(tx)
	if err != nil {
		tx.Rollback()
		return
	}
	err = addShardForUser(tx, userId, shardUrl)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (this *Shards) EnsureShard(shardUrl string) (err error) {
	_, err = this.db.Exec(SqlEnsureShard, shardUrl)
	return
}

// selects shard with the fewest users
func (this *Shards) SelectShard() (shardUrl string, err error) {
	return selectShard(this.db)
}

// selects shard with the fewest users
func selectShard(tx Tx) (shardUrl string, err error) {
	min := MaxInt
	counts, err := getShardUserCount(tx)
	if err != nil {
		return shardUrl, err
	}
	for shard, userCount := range counts {
		if min >= userCount {
			min = userCount
			shardUrl = shard
		}
	}
	if shardUrl == "" {
		err = errors.New("no shard found")
	}
	return
}

func (this *Shards) GetShardUserCount() (result map[string]int, err error) {
	return getShardUserCount(this.db)
}

func getShardUserCount(tx Tx) (result map[string]int, err error) {
	result = map[string]int{}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	rows, err := tx.QueryContext(ctx, SqlShardUserCount)
	if err != nil {
		return
	}
	for rows.Next() {
		var shard string
		var userCount int
		err = rows.Scan(&userCount, &shard)
		if err != nil {
			return result, err
		}
		result[shard] = userCount
	}
	return result, nil
}

func removeShardForUser(tx Tx, userId string) (err error) {
	_, err = tx.Exec(SqlDeleteUserShard, userId)
	return
}

func addShardForUser(tx Tx, userId string, shardAddress string) (err error) {
	_, err = tx.Exec(SqlCreateUserShard, userId, shardAddress)
	return
}

func (this *Shards) GetShards() (result []string, err error) {
	err = this.cache.Use("shards", func() (interface{}, error) {
		return getShards(this.db)
	}, &result)
	return
}

func (this *Shards) RemoveShard(shard string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	tx, err := this.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.Exec(SqlDeleteShardUsers, shard)
	if err != nil {
		return err
	}
	_, err = tx.Exec(SqlDeleteShard, shard)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func getShards(tx Tx) (result []string, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	rows, err := tx.QueryContext(ctx, SQLListShards)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var temp string
		err = rows.Scan(&temp)
		if err != nil {
			return result, err
		}
		result = append(result, temp)
	}
	return result, nil
}
