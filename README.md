# fitbit-readme-stats

![FitBit Heart Rate Chart Example](example.svg)

Plots your heart rate from your FitBit watch in an animated SVG. Embeddable in your [personal profile README.md!](https://docs.github.com/en/github/setting-up-and-managing-your-github-profile/managing-your-profile-readme)
- Heart graphic beats in sync to your current "LIVE" heart rate
- Plots your heart rate from the past 4 hours in a delicious coffee-flavored theme

Inspired by https://github.com/anuraghazra/github-readme-stats

## Setup
1. Execute the latest binary with the `-setup` flag. (`fitbitplot -setup`) on your personal machine.

2. Follow the steps it displays to the terminal.

3. Execute the binary on your desired host (without the `-setup` flag). Include the generated `config.json` in the same directory.

4. Use `![FitBit Heart Rate Chart](http://HOSTIP:8090/stats.svg)` as a README.md embed.
   The SVG is hosted at http://HOSTIP:8090/stats.svg.

## Todo:
- Add releases
- Add config options  
   - Add customizable title
   - Add theming
   - Add plot range
- Add timezone display based on user's timezone (currently uses host's tz)






