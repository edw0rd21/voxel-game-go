#version 410 core

layout (location = 0) in vec2 aPos;    // 2D position in screen space
layout (location = 1) in vec3 aColor;

out vec3 FragColor;

uniform mat4 projection;  // Orthographic projection

void main() {
    gl_Position = projection * vec4(aPos, 0.0, 1.0);
    FragColor = aColor;
}
