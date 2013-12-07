#version 150

in vec2 position;
uniform mat4 projection;
uniform mat4 model;
uniform float scale;

void main()
{
    gl_Position = projection * model * vec4(position*scale, 0.0, 1.0);
}