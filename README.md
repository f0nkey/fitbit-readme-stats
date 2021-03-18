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

## Themes
Replace the `theme` field in your config.json with the codes below.

<details>
<summary>Espresso</summary>

  ```json
"theme": {
   "background": "rgba(50, 35, 35, 255)",
   "text_ticks": "rgba(230, 225, 196, 255)",
   "current_bpm": "rgba(230, 225, 196, 255)",
   "title": "rgba(230, 225, 196, 255)",
   "heart": "rgba(239, 172, 50, 255)",
   "axes": "rgba(239, 93, 50, 255)",
   "plot_line": "rgba(239, 172, 50, 255)",
   "heart_number": "rgba(50, 35, 35, 255)"
},
```
</details>
<img src="./theme-imgs/espresso.svg" alt="espresso theme picture" width=400>

<details>
<summary>GitHub</summary>
 
Uses the "Sponsor" button's pink color.

  ```json
"theme": {
   "background": "rgba(255, 255, 255, 255)",
   "heart_number": "rgba(255, 255, 255, 255)",
   "view_on_github": "rgba(51, 51, 51, 255)",
   "timezone_text": "rgba(51, 51, 51, 255)",
   "text_ticks": "rgba(51, 51, 51, 255)",
   "current_bpm": "rgba(51, 51, 51, 255)",
   "title": "rgba(47, 128, 237, 255)",
   "axes": "rgba(51, 51, 51, 255)",
   "plot_line": "rgba(234, 74, 170, 255)",
   "heart": "rgba(234, 74, 170, 255)"
},
```
</details>
<img src="./theme-imgs/github.svg" alt="github theme picture" width=400>

<details>
<summary>Monokai</summary>

  ```json
"theme": {
   "background": "rgba(39, 40, 34, 255)",
   "heart_number": "rgba(39, 40, 34, 255)",
   "view_on_github": "rgba(226, 137, 5, 255)",
   "timezone_text": "rgba(226, 137, 5, 255)",
   "text_ticks": "rgb(241, 241, 235, 255)",
   "current_bpm": "rgb(241, 241, 235, 255)",
   "title": "rgb(241, 241, 235, 255)",
   "axes": "rgba(226, 137, 5, 255)",
   "plot_line": "rgba(235, 31, 106, 255)",
   "heart": "rgba(235, 31, 106, 255)"
},
```
</details>
<img src="./theme-imgs/monokai.svg" alt="monokai theme picture" width=400>

<details>
<summary>Slate Orange</summary>
   
  ```json
"theme": {
   "background": "rgba(54, 57, 63, 255)",
   "heart_number": "rgba(54, 57, 63, 255)",
   "view_on_github": "rgba(255, 255, 255, 255)",
   "timezone_text": "rgba(255, 255, 255, 255)",
   "text_ticks": "rgba(255, 255, 255, 255)",
   "current_bpm": "rgba(255, 255, 255, 255)",
   "title": "rgba(250, 166, 39, 255)",
   "axes": "rgba(255, 255, 255, 255)",
   "plot_line": "rgba(241, 224, 90, 255)",
   "heart": "rgba(241, 224, 90, 255)"
},
```
</details>
<img src="./theme-imgs/slateorange.svg" alt="slate orange theme picture" width=400>

<details>
<summary>Jolly</summary>

  ```json
"theme": {
   "background": "rgba(41, 27, 62, 255)",
   "heart_number": "rgba(41, 27, 62, 255)",
   "view_on_github": "rgba(255, 255, 255, 255)",
   "timezone_text": "rgba(255, 255, 255, 255)",
   "text_ticks": "rgba(255, 255, 255, 255)",
   "current_bpm": "rgba(255, 255, 255, 255)",
   "title": "rgb(241, 241, 235, 255)",
   "axes": "rgba(169, 96, 255, 255)",
   "plot_line": "rgba(255, 100, 218, 255)",
   "heart": "rgba(255, 100, 218, 255)"
},
```
</details>
<img src="./theme-imgs/jolly.svg" alt="jolly theme picture" width=400>

## Config Documentation
| JSON Field  | Description   |
|-------------|---------------|
| `port` | The port to serve the SVG on. |
| `timezone` | Timezone as an integer hour offset from UTC. Value assumed based on computer's tz during setup. |
| `timezone_abbrev` | The timezone represented in letters e.g., CST, MST. |
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
- More themes?
