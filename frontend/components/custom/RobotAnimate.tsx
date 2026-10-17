"use client";
import React from "react";
import { DotLottieReact } from "@lottiefiles/dotlottie-react";

export const RobotAnimate = () => {
  return (
    <div className="w-[340px] h-[340px] flex items-center justify-center opacity-0 animate-[fadeIn_1.2s_ease-in-out_forwards]">
      <DotLottieReact
        src="/animations/0e15f703-d059-4046-8956-0c93bc391782.json"
        loop
        autoplay
        className="w-full h-full"
      />
    </div>
  );
};
