#version 410 core

layout (location = 0) in vec2 aPos;    // 2D position in screen space
layout (location = 1) in vec3 aColor;
layout (location = 2) in vec2 aTexCoord;

out vec3 FragColor;
out vec2 TexCoord;

uniform mat4 projection;  // Orthographic projection

void main() {
    gl_Position = projection * vec4(aPos, 0.0, 1.0);
    FragColor = aColor;
    TexCoord = aTexCoord;
}
