# Player Sprite Sheet Layout

This document explains how to structure the player sprite sheet files (e.g., `Y.gif`, `R.gif`, `G.gif`, `B.gif`) for the animated movement to work correctly.

## File Structure

Each player sprite file should be a single image containing all the necessary frames for that player's animation. The animation consists of two frames for each of the four cardinal directions.

## Dimensions

-   **Frame Size:** Each individual frame (or cell) in the sprite sheet must be **48px wide by 48px high**.
-   **Total Image Size:** The complete sprite sheet image should be **192px wide** (4 frames across) and **96px high** (2 frames down).

## Frame Layout

The frames are organized in a grid. The horizontal axis (X) represents the player's facing direction, and the vertical axis (Y) represents the animation frame.

The game engine selects which frame to display using CSS `background-position`.

### Visual Grid Layout

Here is a visual representation of how the frames should be arranged in the image file.

```
+---------------------------------------------------------------------------------------+
|                                     (Image Width: 192px)                              |
+----------------------+----------------------+----------------------+----------------------+
|      X=0, Y=0        |     X=-48, Y=0       |     X=-96, Y=0       |    X=-144, Y=0       |
|                      |                      |                      |                      |
|   **Direction: Down**  |  **Direction: Left**   |  **Direction: Right**  |   **Direction: Up**    |
|     (Frame 1)        |     (Frame 1)        |     (Frame 1)        |     (Frame 1)        |
|                      |                      |                      |                      |
+----------------------+----------------------+----------------------+----------------------+ --- (Image Height: 96px)
|      X=0, Y=-48      |    X=-48, Y=-48      |    X=-96, Y=-48      |   X=-144, Y=-48      |
|                      |                      |                      |                      |
|   **Direction: Down**  |  **Direction: Left**   |  **Direction: Right**  |   **Direction: Up**    |
|     (Frame 2)        |     (Frame 2)        |     (Frame 2)        |     (Frame 2)        |
|                      |                      |                      |                      |
+----------------------+----------------------+----------------------+----------------------+

```

### Layout Table

| `background-position` | Direction: Down | Direction: Left | Direction: Right | Direction: Up |
| :--- | :--- | :--- | :--- | :--- |
| **Frame 1** (Y: `0px`) | Down, Frame 1 | Left, Frame 1 | Right, Frame 1 | Up, Frame 1 |
| **Frame 2** (Y: `-48px`) | Down, Frame 2 | Left, Frame 2 | Right, Frame 2 | Up, Frame 2 |
| **X-Position** | `0px` | `-48px` | `-96px` | `-144px` |

### Summary

-   **Top Row (Y=0):** Contains the first animation frame for all four directions.
-   **Bottom Row (Y=-48px):** Contains the second animation frame for all four directions.
-   **Columns (X=0 to -144px):** Correspond to Down, Left, Right, and Up directions, respectively.

When the player is not moving, the `stopped` state defaults to using the first frame (top row) of the current facing direction.
