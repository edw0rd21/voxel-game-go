#version 410 core

in vec3 FragColor;
in vec2 TexCoord;

out vec4 color;

uniform sampler2D uTexture;

void main() {
    vec4 sampled = texture(uTexture, TexCoord);
    color = vec4(FragColor, 1.0) * sampled;
}
