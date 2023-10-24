import pandas as pd
from nba_api.stats.endpoints import commonteamroster
from nba_api.stats.static import teams
from pathlib import Path
import os, json

# get_teams returns a list of 30 dictionaries, each an NBA team.
teamIDs = list()
playerdata = list()
nba_teams = teams.get_teams()
path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/"
print("Number of teams fetched: {}".format(len(nba_teams)))


for team in nba_teams:
    teamIDs.append(team["id"])

blank_playerdata = {
        "player-name" : None,
        "nba-api-pID" : None,
        "nba-api-tID" : None
    } 

for teamID in teamIDs:
    teamfinder = commonteamroster.CommonTeamRoster(season='2023-24',
                                                team_id=str(teamID),
                                                league_id_nullable='00')
    teams = teamfinder.get_data_frames()[0]
    for i in teams.index:
        blank_playerdata["player-name"] = teams["PLAYER"][i]
        blank_playerdata["nba-api-pID"] = int(teams["PLAYER_ID"][i])
        blank_playerdata["nba-api-tID"] = int(teams["TeamID"][i])
        playerdata.append(blank_playerdata)
        blank_playerdata = {
        "player-name" : None,
        "nba-api-pID" : None,
        "nba-api-tID" : None
        } 
print("Number of players fetched: {}".format(len(playerdata)))

path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/"
    
fileName = "playerdata.json"
filePath = os.path.join(path, fileName)

newJSON = json.dumps(playerdata ,indent = 2)

if os.path.exists(path):
    with open(filePath, "w") as f:
        f.write(newJSON)
        

  
