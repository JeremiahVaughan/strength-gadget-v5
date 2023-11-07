import { render } from '@testing-library/react';

import ExercisePage from './exercise-page';

describe('ExercisePage', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ExercisePage />);
    expect(baseElement).toBeTruthy();
  });
});
