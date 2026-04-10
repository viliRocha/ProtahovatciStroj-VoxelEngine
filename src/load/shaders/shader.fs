#version 330
in vec4 fragColor;
in vec3 fragPosition;
in vec3 fragNormal;

uniform vec3 lightDir;

uniform vec4 colDiffuse;

out vec4 finalColor;

uniform vec3 viewPos;
uniform float fogDensity;
uniform float ambient;
uniform vec4 fogColor;

void main() {
    vec3 N = normalize(fragNormal);
    vec3 L = normalize(-lightDir);

    vec3 baseColor = (colDiffuse * fragColor).rgb;

    float diff = max(dot(N, L), 0.2); // never less than 0.2

    // minimum ambient component (so shadows don't turn completely black)
    vec3 litColor = baseColor * (ambient + diff * 0.75);

    // Fog calculation
    float dist = length(viewPos - fragPosition);
    //const vec4 fogColor = vec4(0.588, 0.816, 0.914, 1.0);  // Light Blue
    //const vec4 fogColor = vec4(0.525, 0.051, 0.051, 1.0); Red

    // Linear fog (less nice)
    //const float fogStart = 2.0;
    //const float fogEnd = 10.0;
    //float fogFactor = (fogEnd - dist)/(fogEnd - fogStart);

    // Exponential fog
    float fogFactor = 1.0/exp((dist*fogDensity)*(dist*fogDensity));
    fogFactor = clamp(fogFactor, 0.0, 1.0);

    finalColor = mix(fogColor, vec4(litColor, 1.0), fogFactor);
}