import { render } from '@testing-library/react';

import MeasurementReps from './measurement-reps';

describe('MeasurementReps', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<MeasurementReps />);
    expect(baseElement).toBeTruthy();
  });
});
