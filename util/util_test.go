package util

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

const cidString = "QmSqL792vF4fGjStbYHRgazsEahAQKZmx68jrnvFi9hXMp"

func TestCidToString(t *testing.T) {
	c, err := cid.Decode(cidString)
	require.Nil(t, err)
	s := CidToString(c)
	require.Equal(t, cidString, s)
	c = cid.Undef
	s = CidToString(c)
	require.Equal(t, CidUndef, s)
}

func TestCidFromString(t *testing.T) {
	orig, err := cid.Decode(cidString)
	require.Nil(t, err)
	c, err := CidFromString(cidString)
	require.Nil(t, err)
	require.Equal(t, orig, c)
	c, err = CidFromString(CidUndef)
	require.Nil(t, err)
	require.Equal(t, cid.Undef, c)
	c, err = CidFromString(DefaultCidUndef)
	require.Nil(t, err)
	require.Equal(t, cid.Undef, c)
	_, err = CidFromString("xyz")
	require.NotNil(t, err)
}
