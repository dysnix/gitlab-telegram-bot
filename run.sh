#!/bin/bash

set -e

BOT_API_KEY=${BOT_API_KEY:-}
BOT_HOOK_KEY=${BOT_HOOK_KEY:-}
BOT_ADMIN=${BOT_ADMIN:-}
BOT_LISTEN=${BOT_LISTEN:-"0.0.0.0:3000"}
BOT_DATABASE=${BOT_DATABASE:-"/go/src/app/botdb.sqlite"}


if [[ -z ${BOT_API_KEY} ]]; then
  echo "ERROR: "
  echo "  Please configure the BOT_API_KEY."
  echo "  Cannot continue without a BOT_API_KEY. Aborting..."
  exit 1
fi

if [[ -z ${BOT_HOOK_KEY} ]]; then
  echo "ERROR: "
  echo "  Please configure the BOT_HOOK_KEY."
  echo "  Cannot continue without a BOT_HOOK_KEY. Aborting..."
  exit 1
fi

if [[ -z ${BOT_ADMIN} ]]; then
  echo "ERROR: "
  echo "  Please configure the BOT_ADMIN."
  echo "  Cannot continue without a BOT_ADMIN. Aborting..."
  exit 1
fi

if [ ! -f "/go/src/app/bot.cfg" ]; then
			echo "{\"bot_api\": \"{{BOT_API_KEY}}\",\"hook_key\": \"{{BOT_HOOK_KEY}}\",\"bot_admin\": \"{{BOT_ADMIN}}\",\"listen\": \"{{BOT_LISTEN}}\",\"database\": \"{{BOT_DATABASE}}\"}" > "/go/src/app/bot.cfg"
      sed 's,{{BOT_API_KEY}},'"${BOT_API_KEY}"',g' -i /go/src/app/bot.cfg
      sed 's,{{BOT_HOOK_KEY}},'"${BOT_HOOK_KEY}"',g' -i /go/src/app/bot.cfg
      sed 's,{{BOT_ADMIN}},'"${BOT_ADMIN}"',g' -i /go/src/app/bot.cfg
      sed 's,{{BOT_LISTEN}},'"${BOT_LISTEN}"',g' -i /go/src/app/bot.cfg
      sed 's,{{BOT_DATABASE}},'"${BOT_DATABASE}"',g' -i /go/src/app/bot.cfg
fi



/go/bin/app "/go/src/app/bot.cfg"
