#version 150

out vec4 outColor;
uniform vec3 inColor;

void main()
{
    outColor = vec4(inColor, 1.0);
}