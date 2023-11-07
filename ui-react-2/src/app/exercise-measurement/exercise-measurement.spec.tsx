import { render } from '@testing-library/react';

import ExerciseMeasurement from './exercise-measurement';

describe('ExerciseMeasurement', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ExerciseMeasurement />);
    expect(baseElement).toBeTruthy();
  });
});
