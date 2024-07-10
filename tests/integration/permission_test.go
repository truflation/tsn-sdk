package integration

import (
	"context"
	"fmt"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/core/tsnclient"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"testing"
)

func TestPermissions(t *testing.T) {
	ctx := context.Background()

	// owner assets
	ownerPk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")
	streamOwnerSigner := &auth.EthPersonalSigner{Key: *ownerPk}
	ownerTsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(streamOwnerSigner))
	assertNoErrorOrFail(t, err, "Failed to create client")

	// reader assets
	readerPk, err := crypto.Secp256k1PrivateKeyFromHex("1111111111111111111111111111111111111111111111111111111111111111")
	assertNoErrorOrFail(t, err, "Failed to parse private key")
	readerSigner := &auth.EthPersonalSigner{Key: *readerPk}
	readerAddress, err := util.NewEthereumAddressFromBytes(readerSigner.Identity())
	assertNoErrorOrFail(t, err, "Failed to create signer address")
	readerTsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(readerSigner))
	assertNoErrorOrFail(t, err, "Failed to create client")

	primitiveStreamId := util.GenerateStreamId("test-wallet-permission-primitive-stream")
	composedStreamId := util.GenerateStreamId("test-wallet-permission-composed-stream")

	primitiveStreamLocator := ownerTsnClient.OwnStreamLocator(primitiveStreamId)
	composedStreamLocator := ownerTsnClient.OwnStreamLocator(composedStreamId)

	// destroy the stream at the end
	t.Cleanup(func() {
		destroyResult, err := ownerTsnClient.DestroyStream(ctx, primitiveStreamId)
		assertNoErrorOrFail(t, err, "Failed to destroy stream")
		waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, destroyResult)
	})

	// Deploy a primitive stream
	deployTestPrimitiveStreamWithData(t, ctx, ownerTsnClient, primitiveStreamId, []types.InsertRecordInput{
		{
			Value:     1,
			DateValue: civil.Date{Year: 2020, Month: 1, Day: 1},
		},
	})

	// checkRecords is a helper for asserting the records
	var checkRecords = func(t *testing.T, rec []types.StreamRecord) {
		assert.Equal(t, 1, len(rec))
		assert.Equal(t, "1.000", rec[0].Value.String())
		assert.Equal(t, civil.Date{Year: 2020, Month: 1, Day: 1}, rec[0].DateValue)
	}

	// load primitive stream
	ownerPrimitiveStream, err := ownerTsnClient.LoadPrimitiveStream(primitiveStreamLocator)
	assertNoErrorOrFail(t, err, "Failed to load stream")
	readerPrimitiveStream, err := readerTsnClient.LoadPrimitiveStream(primitiveStreamLocator)
	assertNoErrorOrFail(t, err, "Failed to load stream")

	// will be used to read the stream
	readInput := types.GetRecordsInput{
		DateFrom: &civil.Date{Year: 2020, Month: 1, Day: 1},
		DateTo:   &civil.Date{Year: 2020, Month: 1, Day: 1},
	}

	t.Run("TestPrimitiveStreamWalletReadPermission", func(t *testing.T) {
		t.Cleanup(func() {
			// make these changes not interfere with the next test
			// reset visibility to public
			_, err := ownerPrimitiveStream.SetReadVisibility(ctx, util.PublicVisibility)
			assertNoErrorOrFail(t, err, "Failed to set read visibility")
			// remove permissions
			txHash, err := ownerPrimitiveStream.DisableReadWallet(ctx, readerAddress)
			assertNoErrorOrFail(t, err, "Failed to disable read wallet")

			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash) // only wait the final tx
		})

		// ok - public read
		rec, err := readerPrimitiveStream.GetRecords(ctx, readInput)
		assertNoErrorOrFail(t, err, "Failed to read records")
		checkRecords(t, rec)

		// set the stream to private
		txHash, err := ownerPrimitiveStream.SetReadVisibility(ctx, util.PrivateVisibility)
		assertNoErrorOrFail(t, err, "Failed to set read visibility")
		waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

		// ok - private being owner
		// read the stream
		rec, err = ownerPrimitiveStream.GetRecords(ctx, readInput)
		assertNoErrorOrFail(t, err, "Failed to read records")
		checkRecords(t, rec)

		// fail - private without access
		_, err = readerPrimitiveStream.GetRecords(ctx, readInput)
		assert.Error(t, err)

		// ok - private with access
		// allow read access to the reader
		txHash, err = ownerPrimitiveStream.AllowReadWallet(ctx, readerAddress)
		assertNoErrorOrFail(t, err, "Failed to allow read wallet")
		waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

		// read the stream
		rec, err = readerPrimitiveStream.GetRecords(ctx, readInput)
		assertNoErrorOrFail(t, err, "Failed to read records")
		checkRecords(t, rec)
	})

	t.Run("TestComposedStream", func(t *testing.T) {
		t.Cleanup(func() {
			destroyResult, err := ownerTsnClient.DestroyStream(ctx, composedStreamId)
			assert.NoError(t, err, "Failed to destroy stream")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, destroyResult)
		})

		// preparing the composed stream
		// Deploy a composed stream
		deployTestComposedStreamWithTaxonomy(t, ctx, ownerTsnClient, composedStreamId, []types.TaxonomyItem{
			{
				ChildStream: primitiveStreamLocator,
				Weight:      1,
			},
		})

		// load composed stream
		ownerComposedStream, err := ownerTsnClient.LoadComposedStream(ownerTsnClient.OwnStreamLocator(composedStreamId))
		assertNoErrorOrFail(t, err, "Failed to load stream")
		readerComposedStream, err := readerTsnClient.LoadComposedStream(ownerTsnClient.OwnStreamLocator(composedStreamId))
		assertNoErrorOrFail(t, err, "Failed to load stream")

		t.Run("WalletReadPermission", func(t *testing.T) {
			t.Cleanup(func() {
				// make these changes not interfere with the next test
				// reset visibility to public
				txHash, err := ownerComposedStream.SetReadVisibility(ctx, util.PublicVisibility)
				assert.NoError(t, err, "Failed to set read visibility")

				txHash, err = ownerPrimitiveStream.SetReadVisibility(ctx, util.PublicVisibility)
				assert.NoError(t, err, "Failed to set read visibility")

				// remove permissions from the reader
				txHash, err = ownerComposedStream.DisableReadWallet(ctx, readerAddress)
				assert.NoError(t, err, "Failed to disable read wallet")

				txHash, err = ownerPrimitiveStream.DisableReadWallet(ctx, readerAddress)
				assert.NoError(t, err, "Failed to disable read wallet")

				waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash) // only wait the final tx
			})

			// ok all public
			rec, err := readerComposedStream.GetRecords(ctx, readInput)
			assertNoErrorOrFail(t, err, "Failed to read records")
			checkRecords(t, rec)

			// set just the composed stream to private
			txHash, err := ownerComposedStream.SetReadVisibility(ctx, util.PrivateVisibility)
			assertNoErrorOrFail(t, err, "Failed to set read visibility")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// fail - composed stream is private without access
			_, err = readerComposedStream.GetRecords(ctx, readInput)
			assert.Error(t, err)

			// set the stream to public
			txHash, err = ownerComposedStream.SetReadVisibility(ctx, util.PublicVisibility)
			assertNoErrorOrFail(t, err, "Failed to set read visibility")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// set the child stream to private
			txHash, err = ownerPrimitiveStream.SetReadVisibility(ctx, util.PrivateVisibility)
			assertNoErrorOrFail(t, err, "Failed to set read visibility")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)
			fmt.Println("set private")

			// fail - child is private without access
			_, err = readerComposedStream.GetRecords(ctx, readInput)
			assert.Error(t, err)

			// allow read access to the reader
			txHash, err = ownerPrimitiveStream.AllowReadWallet(ctx, readerAddress)
			assertNoErrorOrFail(t, err, "Failed to allow read wallet")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// ok - primitive private but w/ access
			rec, err = readerComposedStream.GetRecords(ctx, readInput)
			assertNoErrorOrFail(t, err, "Failed to read records")
			checkRecords(t, rec)

			// set the composed stream to private
			txHash, err = ownerComposedStream.SetReadVisibility(ctx, util.PrivateVisibility)
			assertNoErrorOrFail(t, err, "Failed to set read visibility")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// allow read access to the reader
			txHash, err = ownerComposedStream.AllowReadWallet(ctx, readerAddress)
			assertNoErrorOrFail(t, err, "Failed to allow read wallet")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// ok - all private but w/ access
			rec, err = readerComposedStream.GetRecords(ctx, readInput)
			assertNoErrorOrFail(t, err, "Failed to read records")
			checkRecords(t, rec)
		})

		t.Run("StreamComposePermission", func(t *testing.T) {
			t.Cleanup(func() {
				// make these changes not interfere with the next test
				// reset visibility to public
				txHash, err := ownerPrimitiveStream.SetComposeVisibility(ctx, util.PublicVisibility)
				assert.NoError(t, err, "Failed to set compose visibility")
				// remove permissions
				txHash, err = ownerPrimitiveStream.DisableComposeStream(ctx, composedStreamLocator)
				assert.NoError(t, err, "Failed to disable compose stream")

				waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash) // only wait the final tx
			})

			// ok - public compose
			rec, err := readerComposedStream.GetRecords(ctx, readInput)
			assertNoErrorOrFail(t, err, "Failed to read records")
			checkRecords(t, rec)

			// set the stream to private
			txHash, err := ownerPrimitiveStream.SetComposeVisibility(ctx, util.PrivateVisibility)
			assertNoErrorOrFail(t, err, "Failed to set compose visibility")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// ok - reading primitive directly
			rec, err = readerPrimitiveStream.GetRecords(ctx, readInput)
			assertNoErrorOrFail(t, err, "Failed to read records")
			checkRecords(t, rec)

			// fail - private without access
			_, err = readerComposedStream.GetRecords(ctx, readInput)
			assert.Error(t, err)

			// ok - private with access
			// allow compose access to the reader
			txHash, err = ownerPrimitiveStream.AllowComposeStream(ctx, composedStreamLocator)
			assertNoErrorOrFail(t, err, "Failed to allow compose stream")
			waitTxToBeMinedWithSuccess(t, ctx, ownerTsnClient, txHash)

			// read the stream
			rec, err = readerComposedStream.GetRecords(ctx, readInput)
			assertNoErrorOrFail(t, err, "Failed to read records")
			checkRecords(t, rec)
		})
	})

}
