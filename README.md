# SoStat

This code is a bot for Sorare Market (WIP).

When the app is running, the request `/scout` will trigger the bot that will do the following :
- scout player cards among the given teams (here PL teams) which have a lower price than expected
- send a notification on Telegram with the link to the market

To make it run, it needs a`.env` to define your Sorare API token, Telegram bot token and chat ID like so :

    SORARE_API_KEY="YOUR_API_KEY"
    TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN"
    TELEGRAM_CHAT_ID="YOUR_CHAT_ID"
