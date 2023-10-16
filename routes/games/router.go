package games

import (
	"xavagebb/api"
	"xavagebb/routes/games/endpoints/get_available_games"
	"xavagebb/routes/games/endpoints/get_current_game_user"
	"xavagebb/routes/games/endpoints/get_game"
	"xavagebb/routes/games/endpoints/join_game"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Games"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to games"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/users/{userId}/games/{gameId}",
		OpId:    "get_game",
		Method:  uapi.GET,
		Docs:    get_game.Docs,
		Handler: get_game.Route,
		Auth: []uapi.AuthType{
			{
				URLVar:       "userId",
				Type:         api.TargetTypeUser,
				AllowedScope: "notingame", // This endpoint is cross-game
			},
		},
	}.Route(r)

	uapi.Route{
		Pattern: "/users/{userId}/available_games",
		OpId:    "get_available_games",
		Method:  uapi.GET,
		Docs:    get_available_games.Docs,
		Handler: get_available_games.Route,
		Auth: []uapi.AuthType{
			{
				URLVar:       "userId",
				Type:         api.TargetTypeUser,
				AllowedScope: "notingame",
			},
		},
	}.Route(r)

	uapi.Route{
		Pattern: "/users/{userId}/games",
		OpId:    "join_game",
		Method:  uapi.POST,
		Docs:    join_game.Docs,
		Handler: join_game.Route,
		Auth: []uapi.AuthType{
			{
				URLVar:       "userId",
				Type:         api.TargetTypeUser,
				AllowedScope: "notingame",
			},
		},
	}.Route(r)

	uapi.Route{
		Pattern: "/users/{userId}/current_game_user",
		OpId:    "get_current_game_user",
		Method:  uapi.GET,
		Docs:    get_current_game_user.Docs,
		Handler: get_current_game_user.Route,
		Auth: []uapi.AuthType{
			{
				URLVar: "userId",
				Type:   api.TargetTypeUser,
			},
		},
	}.Route(r)
}
