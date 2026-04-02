# PtotahovatsiStroj
### v0.13.4
<center><img src="./img/engine_logo.png" alt="Logo" width="100%"></center>

A voxel engine built using **raylib-go** and **OpenGL**

##  Features 🌟
- **Infinite Random World Generation**: Utilizes a Perlin noise algorithm for creating expansive landscapes.
- **Water Formations**: Realistic water bodies.
- **Surface Feature System**: Procedurally generated trees with [L-systems](https://en.wikipedia.org/wiki/L-system) and randomly placed flowers and tall grass.
- **Cave Generation**: Intricate cave systems made using 3D perlin noise.
- **Biome Diversity**: Various biomes with different topographies determined by Worley noise.
- **Basic Shading**: Combines ambient with directional lighting for better depth perception.
- **Atmospheric effects**: Atmospheric depth with fog and basic clouds.
- **Cache System**: Efficiently stored surface features positions, providing better world consistency.
- **Game Settings**: Configuration menu accessible by pressing "Tab". There players can configure the view distance, FPS limits, world rules (weather, day/night cycle and add/remove or change cloud height), and toggle debug such as FPS and player position.
- **Dynamic Weather**: Weather effects such as rain and snow.

## Upcoming Features 📋
- **Dynamic Time Cycle**: Add a day/night cycle.
- **Advanced Optimizations**: Algorithms such as Fustrum Cooling and Greedy Meshing for better performance.
- **Smooth Lighting**: Improved lighting using vertex lighting and baked ambient occlusion.
- **Block Interaction**: Enable players to place and destroy blocks.
- **Web Build**: Compile the project to WASM.

## Screenshots 🖼️
<img src="./img/lighting3.png" alt="light" width="1000px">
<img src="./img/landscape15.png" alt="world_gen" width="1000px">
<img src="./img/landscape12.png" alt="world_gen" width="1000px">
<img src="./img/landscape16.png" alt="world_gen" width="1000px">

## Getting Started 🚀
To get started with the voxel engine, clone the repository and open the folder. Make sure you have Go installed on your device. Then, run the command `go mod tidy` and finally, to compile the project, run `go run ./src`.

## Controls 🎮
- **Mouse Left Button**: Lock cursor.
- **Camera**: WASD movement, mouse to look.
- **Tab**: Open settings menu.
- **Esc**: To close the window.

## License 📄
This project is licensed under the MIT License - see the `LICENSE` file for details.

## Acknowledgments 🙏
Inspired by [CubeWorld](https://store.steampowered.com/app/1128000/Cube_World/) and [TanTan's](https://github.com/TanTanDev) Voxel engine built with the beavy engine. I would also like to acknowledge the use of voxelized vegetation assets from [MangoVoxel](https://mangovoxel.itch.io/voxelfoliage) in my voxel engine. The shader code was based on examples provided by [raylib](https://www.raylib.com/).<!--Special thanks to the resources and tutorials that helped shape this project.-->