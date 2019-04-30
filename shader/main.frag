#version 450
#extension GL_ARB_separate_shader_objects : enable

layout(location = 0) in vec4 fPos;
layout(location = 1) in vec4 fCol;
layout(location = 2) in vec2 fTex;

layout(binding = 1) uniform sampler2D tex;

layout(location = 0) out vec4 outCol;

vec2 texes[4] = vec2[](
vec2(1.0, 0.0),
vec2(0.0, 1.0),
vec2(0.0, 0.0),
vec2(1.0, 1.0)
);

void main() {
    outCol = texture(tex, fTex);

    float dist = sqrt(max(fPos.z, 1.0)) + 1.0f;

    outCol.rgb *= 0.8f;
    if (fCol.a > 0) {
        outCol.rgb *= 1.1f*fCol.rgb;
        outCol.rgb *= outCol.rgb;
        outCol.rgb /= max(1.0, fPos.z*fPos.z*0.001f);
    }

    outCol.rgb += vec3(0.0, 0.03, 0.03);

    if (outCol.a == 0.0f) {
        discard;
    }
}
