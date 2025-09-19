# DBD Legion Bot

```text

⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⢾⣿⣿⣷⣦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡠⠊⢸⡼⣿⣿⣿⣿⢷⣤⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⣠⣾⣿⣿⣼⣷⣻⣿⣿⣿⣾⢏⢗⣄⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⣠⠺⣦⡻⣿⣿⡿⠋⠁⠀⠉⢻⣟⣷⣿⣿⡷⣄⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣠⣾⣯⢿⣮⣝⣾⢋⠀⠀⠀⠀⠀⠀⣿⣿⢟⣡⣪⣷⣷⣄⠀⠀⠀⠀
⠀⠀⣠⡾⢿⣿⡇⠠⠜⢿⣿⡇⠀⠀⢠⣶⡄⠀⢿⣿⣿⠟⡉⠹⣶⣾⣷⣄⡀⠀
⠀⣾⣷⣶⣮⣤⣿⠀⠸⠟⡍⡁⢤⣴⣿⣿⣷⡤⠠⢟⠻⡇⠀⣴⣿⣿⣿⣿⣿⠆
⠀⠈⠛⢿⢿⣟⣿⠀⢀⣼⡀⠀⠂⠀⣹⣿⡇⠀⣰⣼⣆⠀⢰⣿⣿⣟⡿⠋⠀⠀
⠀⠀⠀⠈⠙⢿⣿⡀⠀⢻⣧⣦⡀⡀⣿⣿⣇⣶⣿⡿⠏⠀⢼⣞⡽⠋⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠈⠙⢷⣶⣾⣿⣿⣷⣸⣿⣿⣿⣿⣿⣧⣤⣬⡿⠋⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⣻⡿⣿⣿⣿⣯⢿⣧⣽⡿⠋⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⢽⣽⣿⣿⡇⣿⠸⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠹⢿⣿⣿⣿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ___  ___ ___    _             _            ___      _
| \| _ )   \ | |   ___ __ _(_)___ _ _   | _ ) ___| |_
| |) | _ \ |) | | |__/ -_) _` | / _ \ ' \ | _ \/ _ \  _|
|___/|___/___/ |____\___\__, |_\___/_||_| |___/\___/\__|
```

A Twitch chatbot that brings Dead by Daylight killer mechanics to your stream's chat experience.
Legion Bot transforms your chat into an interactive game where viewers can be chased, injured, and rescued by various
DbD killers.

# ✨ Features

## 🎮 Interactive Killer Mechanics

    Legion: Frenzy mode where chatters can be hit and need to mend
    Ghost Face: Stalking mechanics with reveal system
    Doctor: Shock therapy that scrambles messages
    Cenobite (Pinhead): Word guessing game with yes/no questions
    Dredge: Realm of darkness with emote-only mode and voting system

## 🌐 Multi-language Support

    English and Russian localization
    Easy to extend with additional languages

## ⚙️ Comprehensive Dashboard

    Web-based control panel for streamers
    Real-time channel status monitoring
    Detailed statistics tracking
    Customizable killer settings and weights

# 🚀 Quick Start

## Prerequisites

    Go 1.24+
    Node.js 18+ (for frontend)
    Twitch Developer Account
    Steam Account (for Steam features)
    Public domain

## Installation

```bash
git clone https://github.com/rofleksey/legion-bot-v2.git
cd legion-bot-v2
cp config.example.yaml config.yaml
cd frontend
npm install
npm run build
cd ..
go mod download
go build -o legion-bot-v2
./legion-bot-v2
```

## Docker Deployment

```bash
docker run -d -p 8080:8080 -v $(pwd)/config.yaml:/opt/config.yaml --name legion-bot legion-bot-v2
```

## 📋 Configuration

See config.example.yaml for example configuration

## 🎯 Usage

Chat Commands
| Command | Description | Example |
| :--- | :--- | :--- |
| `!hp [username]` | Check user's health status | `!hp rofleksey` |
| `!heal [username]` | Heal another user | `!heal @viewer` |
| `!unhook [username]` | Unhook a hooked user | `!unhook @viewer` |
| `!mend` | Mend deep wounds | `!mend` |
| `!pallet` | Attempt to stun Legion | `!pallet` |
| `!locker` | Use Head On ability | `!locker` |
| `!tbag` | Teabag to attract attention | `!tbag` |
| `!reveal` | Attempt to reveal Ghost Face | `!reveal` |
| `!killer` | Show current killer commands | `!killer` |
| `!legiontimeout [duration]` | Temporarily disable bot (streamer only) | `!legiontimeout 1h` |

Killer-Specific Features

Each killer has unique mechanics:

    Legion: Frenzy attacks, pallet stuns, locker grabs
    Ghost Face: Stalking, marking, reveal system
    Doctor: Message scrambling, shock therapy
    Cenobite: Word guessing game with AI responses
    Dredge: Emote-only mode, voting system

# 🙏 Acknowledgments

    Dead by Daylight and all associated killers are property of Behaviour Interactive
    Twitch API for chat and stream integration
    Yandex Cloud for AI capabilities
    The Twitch streaming community for inspiration and testing

* Note: This bot is designed for entertainment purposes and should be used in accordance with Twitch's Terms of Service.
Always obtain proper permissions when interacting with other channels.
