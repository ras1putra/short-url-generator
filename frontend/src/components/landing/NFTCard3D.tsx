"use client";

import { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader.js";

export default function NFTCard3D() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    container.innerHTML = "";

    const scene = new THREE.Scene();

    const ORIGINAL_CAMERA_Z = 4.5;
    const MIN_ZOOM = 2;
    const MAX_ZOOM = 10;
    const RESET_TIMEOUT = 5000;
    const ORIGINAL_ROT_X = 1.5;
    const ORIGINAL_ROT_Y = 1.5;
    let targetCameraZ = ORIGINAL_CAMERA_Z;

    const camera = new THREE.PerspectiveCamera(
      45,
      container.clientWidth / container.clientHeight,
      0.1,
      100
    );
    camera.position.set(0, 0, ORIGINAL_CAMERA_Z);

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.shadowMap.enabled = true;
    container.appendChild(renderer.domElement);

    const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
    scene.add(ambientLight);

    const keyLight = new THREE.DirectionalLight(0xffffff, 1.5);
    keyLight.position.set(2, 3, 4);
    scene.add(keyLight);

    const fillLight = new THREE.DirectionalLight(0x6ee7b7, 0.6);
    fillLight.position.set(-3, 1, 2);
    scene.add(fillLight);

    const rimLight = new THREE.DirectionalLight(0x22d3ee, 0.5);
    rimLight.position.set(0, -2, -3);
    scene.add(rimLight);

    const hologramMaterial = new THREE.ShaderMaterial({
      transparent: true,
      depthWrite: false,
      blending: THREE.AdditiveBlending,
      side: THREE.DoubleSide,
      uniforms: {
        uTime: { value: 0 },
      },
      vertexShader: `
        varying vec3 vNormal;
        varying vec3 vViewPosition;
        varying vec2 vUv;
        void main() {
          vUv = uv;
          vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);
          vNormal = normalize(normalMatrix * normal);
          vViewPosition = -mvPosition.xyz;
          gl_Position = projectionMatrix * mvPosition;
        }
      `,
      fragmentShader: `
        uniform float uTime;
        varying vec3 vNormal;
        varying vec3 vViewPosition;
        varying vec2 vUv;
        void main() {
          vec3 normal = normalize(vNormal);
          vec3 viewDir = normalize(vViewPosition);
          float waveInput = vUv.x * 80.0 - uTime * 2.0;
          float wave = sin(waveInput) * 0.5 + 0.5;
          vec3 waveColor = mix(vec3(0.0), vec3(0.0, 0.64, 1.0), smoothstep(0.0, 0.4, wave));
          float rim = 1.0 - max(dot(normal, viewDir), 0.0);
          rim = pow(rim, 3.0);
          vec3 fresnelColor = mix(vec3(0.0, 0.12, 0.35), vec3(0.0, 1.0, 1.0), smoothstep(0.0, 0.3, rim));
          vec3 emissionColor = mix(waveColor, fresnelColor, 0.6);
          vec3 finalEmission = emissionColor * 2.5;
          float alpha = (0.45 + rim * 0.45) * (0.3 + wave * 0.7);

          gl_FragColor = vec4(finalEmission, alpha);
        }
      `,
    });

    let active = true;

    const loader = new GLTFLoader();
    const modelGroup = new THREE.Group();
    const pivotGroup = new THREE.Group();
    scene.add(pivotGroup);
    pivotGroup.add(modelGroup);

    let animationFrameId: number;

    let isDragging = false;
    const previousMouse = { x: 0, y: 0 };
    let lastInteractionTime = performance.now();
    let isResetting = false;
    let needsReset = false;

    const animate = () => {
      if (!active) return;
      const now = performance.now();
      const timeSinceInteraction = now - lastInteractionTime;

      if (timeSinceInteraction > RESET_TIMEOUT && !isDragging && !isResetting && needsReset) {
        isResetting = true;
        pivotGroup.rotation.y = ((pivotGroup.rotation.y + Math.PI) % (2 * Math.PI)) - Math.PI;
      }

      if (isResetting) {
        const lerpSpeed = 0.05;

        pivotGroup.rotation.x += (0 - pivotGroup.rotation.x) * lerpSpeed;
        pivotGroup.rotation.y += (0 - pivotGroup.rotation.y) * lerpSpeed;

        modelGroup.rotation.x += (ORIGINAL_ROT_X - modelGroup.rotation.x) * lerpSpeed;
        modelGroup.rotation.y += (ORIGINAL_ROT_Y - modelGroup.rotation.y) * lerpSpeed;

        targetCameraZ = ORIGINAL_CAMERA_Z;

        const done =
          Math.abs(pivotGroup.rotation.x) < 0.001 &&
          Math.abs(pivotGroup.rotation.y) < 0.001 &&
          Math.abs(modelGroup.rotation.x - ORIGINAL_ROT_X) < 0.001 &&
          Math.abs(modelGroup.rotation.y - ORIGINAL_ROT_Y) < 0.001 &&
          Math.abs(camera.position.z - ORIGINAL_CAMERA_Z) < 0.01;

        if (done) {
          pivotGroup.rotation.x = 0;
          pivotGroup.rotation.y = 0;
          modelGroup.rotation.x = ORIGINAL_ROT_X;
          modelGroup.rotation.y = ORIGINAL_ROT_Y;
          camera.position.z = ORIGINAL_CAMERA_Z;
          targetCameraZ = ORIGINAL_CAMERA_Z;
          isResetting = false;
          needsReset = false;
          lastInteractionTime = performance.now();
        }
      } else if (!isDragging) {
        pivotGroup.rotation.y += 0.005;
      }

      camera.position.z += (targetCameraZ - camera.position.z) * 0.1;

      hologramMaterial.uniforms.uTime.value += 0.01;
      renderer.render(scene, camera);
      animationFrameId = requestAnimationFrame(animate);
    };

    loader.load(
      "/skip_ads_pass.glb",
      (gltf) => {
        if (!active) return;
        const model = gltf.scene;

        const box = new THREE.Box3().setFromObject(model);
        const center = box.getCenter(new THREE.Vector3());
        const size = box.getSize(new THREE.Vector3());

        const maxDim = Math.max(size.x, size.y, size.z);
        const scale = 3 / maxDim;
        model.scale.setScalar(scale);

        model.position.x = -center.x * scale;
        model.position.y = -center.y * scale;
        model.position.z = -center.z * scale;

        model.traverse((child) => {
          if (child instanceof THREE.Mesh && child.material) {
            if (Array.isArray(child.material)) {
              child.material = child.material.map((mat) =>
                mat.name === "Hologram" ? hologramMaterial : mat
              );
            } else if (child.material.name === "Hologram") {
              child.material = hologramMaterial;
            }
          }
        });

        modelGroup.add(model);

        modelGroup.rotation.x = ORIGINAL_ROT_X;
        modelGroup.rotation.y = ORIGINAL_ROT_Y;

        setLoading(false);
        lastInteractionTime = performance.now();
        animate();
      },
      undefined,
      (err) => {
        if (!active) return;
        console.error("[NFTCard3D] Failed to load 3D GLB model:", err);
        setError("Failed to load 3D Model");
        setLoading(false);
      }
    );

    const onPointerDown = (e: PointerEvent) => {
      e.preventDefault();
      isDragging = true;
      isResetting = false;
      needsReset = true;
      previousMouse.x = e.clientX;
      previousMouse.y = e.clientY;
      lastInteractionTime = performance.now();
      container.setPointerCapture(e.pointerId);
      container.style.cursor = "grabbing";
    };

    const onPointerMove = (e: PointerEvent) => {
      if (!isDragging) return;
      e.preventDefault();
      const deltaX = e.clientX - previousMouse.x;
      const deltaY = e.clientY - previousMouse.y;

      pivotGroup.rotation.y += deltaX * 0.008;
      pivotGroup.rotation.x += deltaY * 0.008;
      pivotGroup.rotation.x = Math.max(-Math.PI / 3, Math.min(Math.PI / 3, pivotGroup.rotation.x));

      previousMouse.x = e.clientX;
      previousMouse.y = e.clientY;
      lastInteractionTime = performance.now();
      needsReset = true;
    };

    const onPointerUp = (e?: PointerEvent) => {
      if (!isDragging) return;
      isDragging = false;
      lastInteractionTime = performance.now();
      container.style.cursor = "grab";
      if (e && container.hasPointerCapture(e.pointerId)) {
        container.releasePointerCapture(e.pointerId);
      }
    };

    const onWheel = (e: WheelEvent) => {
      e.preventDefault();
      e.stopPropagation();

      // Normalize wheel delta across browsers
      let delta = e.deltaY;
      if (e.deltaMode === 1) { // Line mode
        delta *= 40;
      } else if (e.deltaMode === 2) { // Page mode
        delta *= 800;
      }

      targetCameraZ += delta * 0.02;
      targetCameraZ = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, targetCameraZ));
      lastInteractionTime = performance.now();
      isResetting = false;
      needsReset = true;
    };

    const onDragStart = (e: Event) => {
      e.preventDefault();
    };

    container.style.touchAction = "none";
    container.style.cursor = "grab";
    container.addEventListener("pointerdown", onPointerDown);
    container.addEventListener("pointermove", onPointerMove);
    container.addEventListener("pointerup", onPointerUp);
    container.addEventListener("pointercancel", onPointerUp);
    container.addEventListener("pointerleave", onPointerUp);
    container.addEventListener("wheel", onWheel, { passive: false });
    container.addEventListener("dragstart", onDragStart);

    const handleResize = () => {
      if (!container) return;
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };

    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(container);

    return () => {
      active = false;
      cancelAnimationFrame(animationFrameId);
      resizeObserver.disconnect();

      container.style.touchAction = "";
      container.style.cursor = "";
      container.removeEventListener("pointerdown", onPointerDown);
      container.removeEventListener("pointermove", onPointerMove);
      container.removeEventListener("pointerup", onPointerUp);
      container.removeEventListener("pointercancel", onPointerUp);
      container.removeEventListener("pointerleave", onPointerUp);
      container.removeEventListener("wheel", onWheel);
      container.removeEventListener("dragstart", onDragStart);

      scene.traverse((object) => {
        if (!(object instanceof THREE.Mesh)) return;
        object.geometry.dispose();

        if (Array.isArray(object.material)) {
          object.material.forEach((mat) => mat.dispose());
        } else {
          object.material.dispose();
        }
      });

      hologramMaterial.dispose();
      renderer.dispose();

      if (container && renderer.domElement && container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement);
      }
    };
  }, []);

  return (
    <div className="relative w-full h-[500px] md:h-[650px] flex items-center justify-center">
      {loading && (
        <div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-white/[0.01] backdrop-blur-[2px] rounded-2xl border border-white/[0.05]">
          <div className="w-8 h-8 border-2 border-[#6EE7B7]/30 border-t-[#6EE7B7] rounded-full animate-spin mb-2" />
          <span className="text-xs font-mono-dm text-white/40 uppercase tracking-widest">Loading Pass...</span>
        </div>
      )}

      {error && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-black/40 rounded-2xl border border-red-500/20 text-red-400 font-mono-dm text-xs">
          {error}
        </div>
      )}

      <div ref={containerRef} className="w-full h-full" />
    </div>
  );
}