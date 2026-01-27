#version 410 core

in vec3 FragColor;
out vec4 color;

void main() {
    color = vec4(FragColor, 1.0);
}
