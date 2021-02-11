#version 430 core
#define NUMPARTICLES 1000000
layout (local_size_x = 1024, local_size_y = 1) in;
layout (std140, binding = 0) buffer Pos {
  vec4 positions[];
};

layout (std140, binding = 1) buffer Vel {
  vec4 velocities[];
};

layout (std140, binding = 2) buffer Col {
  vec4 colours[];
};

layout (std140, binding = 3) buffer Atr {
  vec4 attractors[];
};


void main() {
  uint index = gl_GlobalInvocationID.x + gl_GlobalInvocationID.y * gl_NumWorkGroups.x * gl_WorkGroupSize.x;

	if(index > NUMPARTICLES) {
    return;
  }

  float t = 0.01, r = 0, g = 0, b = 0;

  vec3 pPos = positions[index].xyz;
  vec3 vPos = velocities[index].xyz;  
  
  for (int i = 0; i < 6; i++) {
    vec3 delta = pPos - attractors[i].xyz;    
    vPos += normalize(delta) * attractors[i].w * t;
    if (i == 0 || i == 3 || i == 4) r += 6/length(delta);
    if (i == 1 || i == 4 || i == 5) g += 6/length(delta);
    if (i == 2 || i == 5 || i == 3) b += 6/length(delta);
  }

  pPos += vPos * t;

  positions[index] = vec4(pPos, 1.0);
  velocities[index] = vec4(vPos, 0.0);  

  colours[index] = vec4(r > 1 ? 1 : r, g > 1 ? 1 : g, b > 1 ? 1 : b, 0.5);  

}