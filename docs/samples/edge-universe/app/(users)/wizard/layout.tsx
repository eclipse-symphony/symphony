"use client";

import Link from 'next/link';
import Box from '@mui/material/Box';
import Stepper from '@mui/material/Stepper';
import Step from '@mui/material/Step';
import StepButton from '@mui/material/StepButton';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import * as React from 'react';
import Step1 from './step1/page';
import Step2 from './step2/page';
import Step3 from './step3/page';
import Stack from '@mui/material/Stack';

const steps = ['Capability Portfolio', 'Site Assets', 'Site Topology', 'Deploy'];

export default function RootLayout ({
    children,
}: {
    children: React.ReactNode;
}) {
    const [activeStep, setActiveStep] = React.useState(0);
    const [completed, setCompleted] = React.useState<{
      [k: number]: boolean;
    }>({});
  
    const totalSteps = () => {
        return steps.length;
      };
    
      const completedSteps = () => {
        return Object.keys(completed).length;
      };
    
      const isLastStep = () => {
        return activeStep === totalSteps() - 1;
      };
    
      const allStepsCompleted = () => {
        return completedSteps() === totalSteps();
      };
    
      const handleNext = () => {
        const newActiveStep =
          isLastStep() && !allStepsCompleted()
            ? // It's the last step, but not all steps have been completed,
              // find the first step that has been completed
              steps.findIndex((step, i) => !(i in completed))
            : activeStep + 1;
        setActiveStep(newActiveStep);
      };
    
      const handleBack = () => {
        setActiveStep((prevActiveStep) => prevActiveStep - 1);
      };
    
      const handleStep = (step: number) => () => {
        setActiveStep(step);
      };
    
      const handleComplete = () => {
        const newCompleted = completed;
        newCompleted[activeStep] = true;
        setCompleted(newCompleted);
        handleNext();
      };
    
      const handleReset = () => {
        setActiveStep(0);
        setCompleted({});
      };

    return (
        <main className="flex">
            <Stack spacing={2}  sx={{ width: '100%' }}>
                
            <Stepper nonLinear activeStep={activeStep} className="w-full">
        {steps.map((label, index) => (
          <Step key={label} completed={completed[index]}>
            <StepButton color="inherit" onClick={handleStep(index)}>
              {label}
            </StepButton>
          </Step>
        ))}
      </Stepper>
      <div className="flex-1">
        {(() => {
          switch (activeStep) {
            case 0:
              return <Step1 />;
            case 1:
              return <Step2 />;
            case 2:
              return <Step3 />;
            default:
              return <div>No Page Found</div>;
          }
        })()}
      </div>
      </Stack>
        </main>
    )
}