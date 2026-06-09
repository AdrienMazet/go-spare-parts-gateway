package db

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedSparePartsCommitsTransaction(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM spare_parts`).WillReturnResult(sqlmock.NewResult(0, 3))
	for range seedSpareParts {
		mock.ExpectExec(`INSERT INTO spare_parts`).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err = SeedSpareParts(db)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSeedSparePartsRollsBackOnInsertError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	expectedErr := errors.New("insert failed")

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM spare_parts`).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`INSERT INTO spare_parts`).WillReturnError(expectedErr)
	mock.ExpectRollback()

	err = SeedSpareParts(db)

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}
