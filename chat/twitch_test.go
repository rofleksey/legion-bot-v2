package chat

import (
	"github.com/nicklaw5/helix/v2"
	"github.com/stretchr/testify/require"
	"legion-bot-v2/config"
	"legion-bot-v2/util"
	"testing"
)

func TestTwitchTimeout(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)

	accessToken, err := util.FetchTwitchAccessToken(cfg.Chat.RefreshToken)
	require.NoError(t, err)

	_, helixClient, err := util.InitTwitchClients(cfg.Chat.ClientID, accessToken)
	require.NoError(t, err)

	botResp, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{util.BotUsername},
	})
	require.NoError(t, err)
	require.Len(t, botResp.Data.Users, 1)

	botUser := botResp.Data.Users[0]

	channelResp, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{"flashbang623"},
	})
	require.NoError(t, err)
	require.Len(t, channelResp.Data.Users, 1)

	channelUser := channelResp.Data.Users[0]

	userResp, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{"rofleksey"},
	})
	require.NoError(t, err)
	require.Len(t, userResp.Data.Users, 1)

	banUser := userResp.Data.Users[0]

	banResp, err := helixClient.BanUser(&helix.BanUserParams{
		BroadcasterID: channelUser.ID,
		ModeratorId:   botUser.ID,
		Body: helix.BanUserRequestBody{
			Duration: 10,
			Reason:   "",
			UserId:   banUser.ID,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "", banResp.ErrorMessage)
}
