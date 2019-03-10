macro_rules! eval {
    ($thunk:expr) => {{
        let thunk = $thunk;
        (thunk.entry)(&mut thunk.payload)
    }};
    ($thunk:expr, $($arg:expr),+) => {{
        let thunk = &mut $thunk;
        (thunk.entry)(&mut thunk.payload, $($arg),+)
    }};
}

macro_rules! closure {
    ($result:ty) => {
        crate::core::Closure<extern "fastcall" fn(&mut $result) -> &mut $result, $result>
    };
    ($result:ty, $($arg:ty),+) => {
        crate::core::Closure<extern "fastcall" fn(&mut crate::core::Environment, $($arg),+) -> $result, crate::core::Environment>
    };
}

#[derive(Clone, Copy)]
#[repr(C)]
pub struct Closure<E, P> {
    pub entry: E,
    pub payload: P,
}

impl<E, P> Closure<E, P> {
    pub fn new(entry: E, payload: P) -> Self {
        Closure { entry, payload }
    }
}

impl<T: Clone> From<&[T]> for List<T> {
    fn from(xs: &[T]) -> List<T> {
        Closure::new(list_entry, xs.into())
    }
}

impl From<f64> for Number {
    fn from(n: f64) -> Number {
        Closure::new(number_entry, n.into())
    }
}

#[repr(C)]
pub struct Environment {
    _private: [u8; 0],
}

pub type List<T> = closure!(algebraic::List<T>);
pub type Number = closure!(algebraic::Number);

extern "fastcall" fn list_entry<T: Clone>(
    list: &mut algebraic::List<T>,
) -> &mut algebraic::List<T> {
    list
}

extern "fastcall" fn number_entry(number: &mut algebraic::Number) -> &mut algebraic::Number {
    number
}

pub mod algebraic {
    #[derive(Clone, Copy)]
    #[repr(C)]
    pub enum List<T: Clone> {
        Cons(*mut T, *mut super::List<T>),
        Nil,
    }

    impl<T: Clone, S: Clone + Into<T>> From<&[S]> for List<T> {
        fn from(xs: &[S]) -> List<T> {
            let mut l = List::Nil;

            for x in xs.into_iter().rev() {
                l = List::Cons(
                    &mut x.clone().into(),
                    &mut super::Closure::new(super::list_entry, l),
                )
            }

            l
        }
    }

    #[derive(Clone, Copy)]
    #[repr(C)]
    pub struct Number(f64);

    impl From<f64> for Number {
        fn from(n: f64) -> Number {
            Number(n)
        }
    }

    impl From<Number> for f64 {
        fn from(n: Number) -> f64 {
            n.0
        }
    }
}
