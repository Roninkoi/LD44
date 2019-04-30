#version 450
#extension GL_ARB_separate_shader_objects : enable

layout(binding = 0) uniform UniformBufferObject {
    mat4 obj;
    mat4 cam;
    mat4 proj;
    vec4 amb;
} ubo;

layout(location = 0) in vec4 pos;
layout(location = 1) in vec4 col;
layout(location = 2) in vec4 tex;

layout(location = 0) out vec4 fPos;
layout(location = 1) out vec4 fCol;
layout(location = 2) out vec2 fTex;

void main() {
    gl_Position = ubo.obj * vec4(pos.xyz, 1.0);
    if (pos.w > 0.0) {
        gl_Position = ubo.proj * ubo.cam * gl_Position;
    }
    fPos = gl_Position;
    fCol = col;
    fCol.rgb *= ubo.amb.rgb;
    fCol.a = pos.w;
    fTex = vec2(tex.x*tex.z, tex.y*tex.w);
}
