#version 430

in layout(location = 0) vec4 position;

layout(location = 1) uniform mat4 projMat;
layout(location = 2) uniform mat4 viewMat;

in layout(location = 3) vec4 colour;

out vec4 pointColour;

void main() {
  gl_Position = projMat * viewMat * position;
  pointColour = colour;
}