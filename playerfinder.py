import shutil
import pandas as pd
from nba_api.stats.endpoints import commonteamroster
from nba_api.stats.static import teams
from pathlib import Path
import os, json

teamIDs = list()
playerdata = list()
nba_teams = teams.get_teams() # get_teams returns a list of 30 dictionaries, each an NBA team.
path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_data/"
print("Number of teams fetched: {}".format(len(nba_teams)))

problem_players = [ 
    'Miles Norris', 'Trent Forrest', 'Garrison Mathews', 'Nathan Knight', 'Dalano Banton', 'Lamar Stevens',
    'Max Strus', 'Sam Merrill', 'Craig Porter', 'Dean Wade', 'Kaiser Gates', 'Terry Taylor', 'Justin Lewis',
    'Greg Brown III', 'Dexter Dennis', 'Jermaine Samuels Jr.', "Jae'Sean Tate", 'Trevor Hudgins',
    'Reggie Bullock Jr.', 'Nate Williams', 'Colin Castleton', 'Alex Fudge', "D'Moi Hodge", 'Jamal Cain',
    'Dru Smith', 'Lindell Wigginton', 'Armoni Brooks', 'Nic Claxton', 'Jacob Toppin', 'DaQuan Jeffries',
    'Charlie Brown Jr.', 'Ben Sheppard', 'Filip Petrusev', 'Jordan Goodwin', 'Justin Minaya', 'Skylar Mays',
    'Ish Wainright', 'Ibou Badji', 'Jalen Slawson', 'Kessler Edwards', 'Charles Bediako', 'Lindy Waters III',
    'Kenrich Williams', 'Javon Freeman-Liberty', 'Micah Potter', 'Jacob Gilyard', 'Vince Williams Jr.',
    'Kenneth Lofton Jr.', 'Xavier Cooks', 'Eugene Omoruyi', 'Jared Rhoden', 'Stanley Umude', 'Leaky Black',
    'JT Thor', 'Darius Days', 'Jeremiah Robinson-Earl', "Duop Reath"
] #list of players that either have no results from the getty image webscrape/don't have an image recognizable by the go-face package 


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
        if teams["PLAYER"][i] in problem_players:
            continue
        else:
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

path = "/mnt/c/Users/kodar/Documents/CS-Work/NBA-Player-TwinApp/player_headshots/" 

for player in problem_players:
    folder_path = os.path.join(path, player)
    if os.path.exists(folder_path):
        shutil.rmtree(folder_path)
        print(f"Deleted folder for {player}")
    else:
        print(f"No folder found for {player}. Skipping.")