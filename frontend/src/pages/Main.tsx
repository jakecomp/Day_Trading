import { Routes, Route } from 'react-router-dom'
import { Home } from './Home'
import { SignIn } from './SignIn'
import { SignUp } from './SignUp'

const Main = () => (
    <Routes>
        <Route
            path='/'
            element={
                <>
                    <SignUp />
                </>
            }
        ></Route>
        <Route
            path='/signin'
            element={
                <>
                    <SignIn />
                </>
            }
        ></Route>
        <Route
            path='/home'
            element={
                <>
                    <Home />
                </>
            }
        ></Route>
    </Routes>
)
export default Main
