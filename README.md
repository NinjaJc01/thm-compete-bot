# thm-compete-bot
## This is currently broken due to TryHackMe using cloudflare to help deal with the Advent Of Cyber load.
Discord Bot to encourage competition on TryHackMe between my housemates.
Prints a leaderboard at 10PM daily, and allows retrieval of user stats
## Config file
The bot expects a config file in the CWD when it runs.
The structure is as follows:
```json
{
    "token":"PutYourDiscordTokenHere",
    "users":["NinjaJc01","szymex73"],
    "prefix":"£",
    "leaderboardTime": "00 22 * * *",
    "channelID":"DesiredChannelIDForDailyStatsGoesHere"
}
```
## Commands - Replace £ with the prefix in your config file
- £stats NinjaJc01 - Retrieve the stats for THM User NinjaJc01
