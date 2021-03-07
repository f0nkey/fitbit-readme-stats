# fitbit-readme-stats

![FitBit Heart Rate Chart Example](example.svg)

Plots your heart rate from your FitBit watch in an animated SVG. Embeddable in your [personal profile README.md!](https://docs.github.com/en/github/setting-up-and-managing-your-github-profile/managing-your-profile-readme)
- Heart graphic beats in sync to your current "LIVE" heart rate
- Plots your heart rate from the past 4 hours in a delicious coffee-flavored theme

Inspired by https://github.com/anuraghazra/github-readme-stats

## Setup
1. [Execute the latest binary](https://github.com/f0nkey/fitbit-readme-stats/releases) with the `-setup` flag. (`fitbitplot -setup`) on your personal machine.

2. Follow the steps it displays to the terminal to generate `config.json`.

3. Execute the binary on your desired host (without the `-setup` flag). Include the generated `config.json` in the same directory.

4. Use `![FitBit Heart Rate Chart](http://HOSTIP:8090/stats.svg)` as a README.md embed.
   The SVG is hosted at http://HOSTIP:8090/stats.svg.

## Config Documentation
| JSON Field  | Description   |
|-------------|---------------|
| `port` | The port to serve the SVG on. |
| `banner_title` | The title at the top of the banner. |
| `cache_invalidation_time` | How long (in seconds) before new heart-rate data should be requested from FitBit's servers. Checked every SVG request. |
| `plot_range` | The time interval (in hours) to look back for heart-rate data. |
| `banner_width` | The width of the generated .SVG. |
| `banner_height` | The height of the generated .SVG. |
| `display_view_on_github` | When true, displays watermark/link to this GitHub repo in the top left. |
| `theme` | Colors for each element. Represented as: `rgba(255, 255, 255, 255)` |
| `app_credentials` | Holds generated fields when a new app is made at https://dev.fitbit.com/. |
| `user_credentials` | Holds credentials to authenticate with and request from the FitBit Web API. Don't share it with anyone! |

## Todo
- Add timezone display based on user's timezone (currently uses host's tz)
