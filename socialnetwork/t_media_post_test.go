package socialnetwork_test

import (
	"github.com/stretchr/testify/assert"
	"sigmaos/fslib"
	"sigmaos/rpcclnt"
	sn "sigmaos/socialnetwork"
	"sigmaos/socialnetwork/proto"
	"sigmaos/test"
	"testing"
)

func IsPostEqual(a, b *proto.Post) bool {
	if a.Postid != b.Postid || a.Posttype != b.Posttype ||
		a.Timestamp != b.Timestamp || a.Text != b.Text ||
		a.Creator != b.Creator || len(a.Medias) != len(a.Medias) ||
		len(a.Usermentions) != len(b.Usermentions) || len(a.Urls) != len(b.Urls) {
		return false
	}
	for idx, _ := range a.Usermentions {
		if a.Usermentions[idx] != b.Usermentions[idx] {
			return false
		}
	}
	for idx, _ := range a.Urls {
		if a.Urls[idx] != b.Urls[idx] {
			return false
		}
	}
	for idx, _ := range a.Medias {
		if a.Medias[idx] != b.Medias[idx] {
			return false
		}
	}
	return true
}

func TestMedia(t *testing.T) {
	// start server
	tssn := makeTstateSN(t, []sn.Srv{sn.Srv{"socialnetwork-media", test.Overlays, 1000}}, NCACHESRV)
	snCfg := tssn.snCfg

	// create a RPC client and query
	rpcc, err := rpcclnt.MkRPCClnt([]*fslib.FsLib{snCfg.FsLib}, sn.SOCIAL_NETWORK_MEDIA)
	assert.Nil(t, err, "RPC client should be created properly")

	// store two media
	mdata1 := []byte{1, 3, 5, 7, 9, 11, 13, 15}
	mdata2 := []byte{2, 3, 5, 7, 11, 13, 17, 19}
	arg_store := proto.StoreMediaRequest{Mediatype: "File", Mediadata: mdata1}
	res_store := proto.StoreMediaResponse{}
	assert.Nil(t, rpcc.RPC("MediaSrv.StoreMedia", &arg_store, &res_store))
	assert.Equal(t, "OK", res_store.Ok)
	mId1 := res_store.Mediaid
	arg_store = proto.StoreMediaRequest{Mediatype: "Video", Mediadata: mdata2}
	assert.Nil(t, rpcc.RPC("MediaSrv.StoreMedia", &arg_store, &res_store))
	assert.Equal(t, "OK", res_store.Ok)
	mId2 := res_store.Mediaid

	// read the medias
	arg_read := proto.ReadMediaRequest{Mediaids: []int64{mId1, mId2}}
	res_read := proto.ReadMediaResponse{}
	assert.Nil(t, rpcc.RPC("MediaSrv.ReadMedia", &arg_read, &res_read))
	assert.Equal(t, "OK", res_read.Ok)
	assert.Equal(t, 2, len(res_read.Mediatypes))
	assert.Equal(t, 2, len(res_read.Mediadatas))
	assert.Equal(t, "File", res_read.Mediatypes[0])
	assert.Equal(t, "Video", res_read.Mediatypes[1])
	assert.Equal(t, mdata1, res_read.Mediadatas[0])
	assert.Equal(t, mdata2, res_read.Mediadatas[1])

	// stop server
	assert.Nil(t, tssn.Shutdown())
}

func TestPost(t *testing.T) {
	// start server
	tssn := makeTstateSN(t, []sn.Srv{sn.Srv{"socialnetwork-post", test.Overlays, 1000}}, NCACHESRV)
	snCfg := tssn.snCfg

	// create a RPC client and query
	rpcc, err := rpcclnt.MkRPCClnt([]*fslib.FsLib{snCfg.FsLib}, sn.SOCIAL_NETWORK_POST)
	assert.Nil(t, err, "RPC client should be created properly")

	// create two posts
	post1 := proto.Post{
		Postid:       int64(1),
		Posttype:     proto.POST_TYPE_POST,
		Timestamp:    int64(12345),
		Creator:      int64(200),
		Text:         "First Post",
		Usermentions: []int64{int64(201)},
		Medias:       []int64{int64(777)},
		Urls:         []string{"XXXXX"},
	}
	post2 := proto.Post{
		Postid:       int64(2),
		Posttype:     proto.POST_TYPE_REPOST,
		Timestamp:    int64(67890),
		Creator:      int64(200),
		Text:         "Second Post",
		Usermentions: []int64{int64(202)},
		Urls:         []string{"YYYYY"},
	}

	// store first post
	arg_store := proto.StorePostRequest{Post: &post1}
	res_store := proto.StorePostResponse{}
	assert.Nil(t, rpcc.RPC("PostSrv.StorePost", &arg_store, &res_store))
	assert.Equal(t, "OK", res_store.Ok)

	// check for two posts. one missing
	arg_read := proto.ReadPostsRequest{Postids: []int64{int64(1), int64(2)}}
	res_read := proto.ReadPostsResponse{}
	assert.Nil(t, rpcc.RPC("PostSrv.ReadPosts", &arg_read, &res_read))
	assert.Equal(t, "No. Missing 2.", res_read.Ok)

	// store second post and check for both.
	arg_store.Post = &post2
	assert.Nil(t, rpcc.RPC("PostSrv.StorePost", &arg_store, &res_store))
	assert.Equal(t, "OK", res_store.Ok)
	assert.Nil(t, rpcc.RPC("PostSrv.ReadPosts", &arg_read, &res_read))
	assert.Equal(t, "OK", res_read.Ok)
	assert.True(t, IsPostEqual(&post1, res_read.Posts[0]))
	assert.True(t, IsPostEqual(&post2, res_read.Posts[1]))

	//stop server
	assert.Nil(t, tssn.Shutdown())
}
