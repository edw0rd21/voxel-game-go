#version 410 core

out vec4 FragColor;

in vec2 TexCoord;
in vec3 Normal;
in vec3 FragPos;

uniform sampler2D texture1;
uniform vec3 lightDir;

void main() {
    vec4 texColor = texture(texture1, TexCoord);

    vec3 norm = normalize(Normal);
    vec3 lightDirNormalized = normalize(-lightDir);
    float diff = max(dot(norm, lightDirNormalized), 0.0);
    
    vec3 ambient = 0.3 * texColor.rgb;
    vec3 diffuse = diff * texColor.rgb;
    
    FragColor = vec4(ambient + diffuse, 1.0);
}
