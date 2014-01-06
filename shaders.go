package main

const (
	block_fragment_shader = `
		#version 150

		out vec4 outColor;
		uniform vec3 inColor;

		void main()
		{
		    outColor = vec4(inColor, 1.0);
		}
	`

	block_vertex_shader = `
		#version 150

		in vec2 position;
		uniform mat4 projection;
		uniform mat4 model;
		uniform float scale;
		uniform bool reflect;

		void main()
		{
			if (reflect) {
				position.x = -position.x;
			}
		    gl_Position = projection * model * vec4(position*scale, 0.0, 1.0);
		}
	`
)
